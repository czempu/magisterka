from pyscipopt import Model, quicksum
import time
from core.models import ScheduleRequest, ScheduleResponse
from core.calendar_utils import CalendarContext
from core.shift_rules import calculate_forbidden_transitions

def solve_with_scip(request: ScheduleRequest) -> ScheduleResponse:
    start_time = time.perf_counter()
    model = Model("Schedule")
    model.hideOutput()
    model.setRealParam("limits/time", request.time_limit_seconds)
    
    shifts = request.shifts
    num_shifts = len(shifts)
    shift_duration = {s.id: s.duration_hours for s in shifts}
    base_hours = shifts[0].duration_hours if shifts else 8

    # 0. KONTEKST KALENDARZOWY
    cal_ctx = None
    calendar_dates = []
    holidays_str = []
    closed_days_str = []
    base_norm_hours = request.max_work_hours
    
    if request.start_date and request.end_date:
        cal_ctx = CalendarContext(
            start_date=request.start_date,
            end_date=request.end_date,
            custom_holidays=request.custom_holidays,
            closed_days=request.closed_days,
            shift_hours=base_hours
        )
        calendar_dates = [cal_ctx.get_date_str(d) for d in range(cal_ctx.num_days)]
        holidays_str = [d.strftime("%Y-%m-%d") for d in cal_ctx.holidays]
        closed_days_str = [d.strftime("%Y-%m-%d") for d in cal_ctx.closed_days]
        
        if request.auto_calculate_work_norm and base_norm_hours is None:
            base_norm_hours = cal_ctx.calculate_kp_work_norm_hours()

    if base_norm_hours is None:
        base_norm_hours = int((request.num_days / 7) * 40)

    # 1. ZMIENNE DECYZYJNE: x[e, d, s] in {0, 1}
    x = {}
    for e in range(request.num_employees):
        for d in range(request.num_days):
            for s in range(num_shifts):
                x[(e, d, s)] = model.addVar(vtype="B", name=f"x_e{e}_d{d}_s{s}")
                
    # 2. REGUŁA TWARDA: Max 1 zmiana dziennie dla pracownika
    for e in range(request.num_employees):
        for d in range(request.num_days):
            model.addCons(quicksum(x[(e, d, s)] for s in range(num_shifts)) <= 1)

    # 3. REGUŁA TWARDA: Dynamiczne 11h odpoczynku dobowego
    forbidden_transitions = calculate_forbidden_transitions(shifts, min_rest_hours=11)
    for (s1, s2) in forbidden_transitions:
        for e in range(request.num_employees):
            for d in range(request.num_days - 1):
                model.addCons(x[(e, d, s1)] + x[(e, d + 1, s2)] <= 1)
            
    # 4. REGUŁA TWARDA: Odpoczynek tygodniowy 35h
    if request.num_days >= 7:
        for e in range(request.num_employees):
            for start_d in range(request.num_days - 6):
                model.addCons(
                    quicksum(x[(e, start_d + k, s)] for k in range(7) for s in range(num_shifts))
                    <= request.max_consecutive_days
                )

    # 5. REGUŁA TWARDA: Co 4. niedziela wolna od pracy
    if request.require_sunday_rest_every_4_weeks and cal_ctx:
        sunday_indices = [d for d in range(cal_ctx.num_days) if cal_ctx.is_sunday(d)]
        if len(sunday_indices) >= 4:
            for e in range(request.num_employees):
                for w in range(len(sunday_indices) - 3):
                    four_sundays = [sunday_indices[w + k] for k in range(4)]
                    model.addCons(
                        quicksum(x[(e, sun_d, s)] for sun_d in four_sundays for s in range(num_shifts)) <= 3
                    )

    # 6. REGUŁA TWARDA: Zapotrzebowanie (Demand) oraz Dni Zamknięte
    for d in range(request.num_days):
        is_closed = cal_ctx.is_closed(d) if cal_ctx else False
        for s in range(num_shifts):
            if is_closed:
                model.addCons(quicksum(x[(e, d, s)] for e in range(request.num_employees)) == 0)
            else:
                required = request.demand.get(f"d{d}_s{s}", 1)
                model.addCons(quicksum(x[(e, d, s)] for e in range(request.num_employees)) == required)

    # 6b. REGUŁA TWARDA: Zapotrzebowanie na Kwalifikacje / Stanowiska (Skill Demand)
    if request.skill_demand:
        for d in range(request.num_days):
            is_closed = cal_ctx.is_closed(d) if cal_ctx else False
            if not is_closed:
                for s in range(num_shifts):
                    for (key, min_count) in request.skill_demand.items():
                        match = False
                        skill_name = None
                        prefix_d_s = f"d{d}_s{s}_k_"
                        prefix_s = f"s{s}_k_"

                        if key.startswith(prefix_d_s):
                            skill_name = key[len(prefix_d_s):]
                            match = True
                        elif key.startswith(prefix_s) and not any(k.startswith(prefix_d_s) for k in request.skill_demand):
                            skill_name = key[len(prefix_s):]
                            match = True
                        elif key.startswith("k_") and not any(k.startswith(prefix_s) or k.startswith(prefix_d_s) for k in request.skill_demand):
                            skill_name = key[2:]
                            match = True

                        if match and skill_name:
                            skilled_emps = [e_idx for e_idx, emp in enumerate(request.employees) if skill_name in emp.skills]
                            model.addCons(quicksum(x[(e_idx, d, s)] for e_idx in skilled_emps) >= min_count)

    # 7. PROFIL PRACOWNIKA: Urlopy, Zakazy zmian oraz Indywidualny Wymiar Etatu (FTE)
    for e_idx, emp in enumerate(request.employees):
        emp_max_hours = int(base_norm_hours * emp.contract_fte)
        emp_hours_expr = quicksum(x[(e_idx, d, s)] * shift_duration[s] for d in range(request.num_days) for s in range(num_shifts))
        model.addCons(emp_hours_expr <= emp_max_hours)

        if cal_ctx and emp.vacation_days:
            for vac_date in emp.vacation_days:
                if vac_date in cal_ctx.dates:
                    d_idx = cal_ctx.dates.index(vac_date)
                    model.addCons(quicksum(x[(e_idx, d_idx, s)] for s in range(num_shifts)) == 0)

        for dis_s in emp.disallowed_shift_ids:
            if 0 <= dis_s < num_shifts:
                for d in range(request.num_days):
                    model.addCons(x[(e_idx, d, dis_s)] == 0)

    # 8. FUNKCJA CELU (Reguły miękkie):
    objective_expr = []
    for (key, penalty) in request.shift_penalties.items():
        parts = key.split('_')
        if len(parts) == 3:
            e = int(parts[0][1:])
            d = int(parts[1][1:])
            s = int(parts[2][1:])
            if (e, d, s) in x:
                objective_expr.append(penalty * x[(e, d, s)])

    # Preferencje z profili pracowników
    for e_idx, emp in enumerate(request.employees):
        for pref in emp.preferences:
            target_d = None
            if pref.day_index is not None and 0 <= pref.day_index < request.num_days:
                target_d = pref.day_index
            elif cal_ctx and pref.pref_date and pref.pref_date in cal_ctx.dates:
                target_d = cal_ctx.dates.index(pref.pref_date)

            if target_d is not None:
                if pref.want_free_day:
                    work_d_expr = quicksum(x[(e_idx, target_d, s)] for s in range(num_shifts))
                    objective_expr.append(pref.weight * work_d_expr)
                elif pref.preferred_shift_id is not None:
                    for s in range(num_shifts):
                        if s != pref.preferred_shift_id:
                            objective_expr.append(pref.weight * x[(e_idx, target_d, s)])

    # Rotacja wsteczna
    if request.enforce_forward_rotation:
        for e in range(request.num_employees):
            for d in range(request.num_days - 1):
                for s1 in range(num_shifts):
                    for s2 in range(num_shifts):
                        if shifts[s2].start_hour < shifts[s1].start_hour:
                            rot_pen_var = model.addVar(vtype="B", name=f"rot_{e}_{d}_{s1}_{s2}")
                            model.addCons(rot_pen_var >= x[(e, d, s1)] + x[(e, d + 1, s2)] - 1)
                            objective_expr.append(5 * rot_pen_var)

    if cal_ctx:
        for e in range(request.num_employees):
            for d in range(request.num_days):
                if cal_ctx.is_holiday(d) or cal_ctx.is_sunday(d):
                    work_d = quicksum(x[(e, d, s)] for s in range(num_shifts))
                    objective_expr.append(1 * work_d)

    if objective_expr:
        model.setObjective(quicksum(objective_expr), "minimize")

    # ROZWIĄZYWANIE
    model.optimize()
    solve_time = time.perf_counter() - start_time
    
    result_schedule = {}
    assigned_hours = {f"emp_{e}": 0 for e in range(request.num_employees)}
    employee_names = {f"emp_{e.id}": e.name for e in request.employees}
    status_name = model.getStatus()
    obj_val = None
    
    if status_name in ("optimal", "feasible"):
        if objective_expr:
            obj_val = model.getObjVal()
        else:
            obj_val = 0.0

        for e in range(request.num_employees):
            total_emp_hours = 0
            for d in range(request.num_days):
                for s in range(num_shifts):
                    if model.getVal(x[(e, d, s)]) > 0.5:
                        result_schedule[f"emp_{e}_day_{d}"] = s
                        total_emp_hours += shift_duration[s]
            assigned_hours[f"emp_{e}"] = total_emp_hours
                        
    employee_skills = {f"emp_{e.id}": e.skills for e in request.employees}
    
    return ScheduleResponse(
        solver_name="SCIP",
        status=status_name.upper(),
        solve_time_seconds=solve_time,
        objective_value=obj_val,
        schedule=result_schedule,
        assigned_hours=assigned_hours,
        employee_names=employee_names,
        employee_skills=employee_skills,
        shift_definitions=shifts,
        calendar_dates=calendar_dates,
        calculated_work_norm_hours=base_norm_hours,
        holidays=holidays_str,
        closed_days=closed_days_str
    )