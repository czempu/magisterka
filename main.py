from datetime import date
from core.models import ScheduleRequest
from solvers.cp_solver import solve_with_cpsat
from solvers.scip_solver import solve_with_scip
from core.calendar_utils import CalendarContext

def print_calendar_schedule(schedule_dict, num_employees, cal_ctx: CalendarContext, assigned_hours=None):
    # Nagłówek dni tygodnia
    row_weekdays = "      "
    row_dates = "      "
    row_sep = "      "
    for d in range(cal_ctx.num_days):
        wd = cal_ctx.get_weekday_pl(d)
        d_num = cal_ctx.dates[d].day
        is_hol = cal_ctx.is_holiday(d)
        is_sun = cal_ctx.is_sunday(d)
        is_cls = cal_ctx.is_closed(d)
        
        tag = ""
        if is_cls:
            tag = "X"
        elif is_hol:
            tag = "*"
        elif is_sun:
            tag = "n"
            
        header_cell = f"{wd:<2}{tag:<2}"
        row_weekdays += f"{header_cell:<5}"
        row_dates += f"{d_num:>2}   "
        row_sep += "-----"
        
    if assigned_hours:
        row_weekdays += " | Godziny"
        row_dates += " | Etatu"
        row_sep += " | --------"
        
    print("\n" + row_weekdays)
    print(row_dates)
    print(row_sep)
    
    for e in range(num_employees):
        row = f"Prac{e:<2}|"
        for d in range(cal_ctx.num_days):
            shift = schedule_dict.get(f"emp_{e}_day_{d}")
            if shift is not None:
                row += f"  {shift}  "
            else:
                if cal_ctx.is_closed(d):
                    row += " [X] "
                else:
                    row += "  -  "
        if assigned_hours:
            hours = assigned_hours.get(f"emp_{e}", 0)
            row += f" |  {hours}h"
        print(row)
    print(row_sep + "\n")
    print("Legenda: [*] Święto ustawowe, [n] Niedziela, [X] Dzień zamknięty firmy\n")

def main():
    print("==================================================================")
    print("   MAGISTERKA: HARMONOGRAMOWANIE KALENDARZOWE Z KODEKSEM PRACY")
    print("==================================================================\n")
    
    # Przełom miesięcy: 25 Października 2026 -> 15 Listopada 2026 (22 dni)
    start_date = date(2026, 10, 25)
    end_date = date(2026, 11, 15)
    
    # Dodatkowe własne święto firmowe: 2 listopada (Dzień Zaduszny jako wolne w firmie)
    custom_holidays = [date(2026, 11, 2)]
    
    # Dzień całkowitego zamknięcia zakładu: 1 listopada (Wszystkich Świętych - zakaz handlu / prac)
    closed_days = [date(2026, 11, 1)]
    
    cal_ctx = CalendarContext(
        start_date=start_date,
        end_date=end_date,
        custom_holidays=custom_holidays,
        closed_days=closed_days,
        shift_hours=8
    )
    
    kp_norm = cal_ctx.calculate_kp_work_norm_hours()
    
    print(f"Zakres planowania: {start_date} do {end_date} (Dni: {cal_ctx.num_days})")
    print(f"Święta w okresie: {[d.strftime('%Y-%m-%d') for d in cal_ctx.holidays]}")
    print(f"Dni zamknięte zakładu: {[d.strftime('%Y-%m-%d') for d in cal_ctx.closed_days]}")
    print(f"Wyliczona norma godzinowa (art. 130 KP): {kp_norm} godzin na etat\n")
    
    # 8 pracowników, 3 zmiany
    num_employees = 8
    num_shifts = 3
    
    # Zapotrzebowanie: 1 osoba na zmianę
    demand = {}
    for d in range(cal_ctx.num_days):
        for s in range(num_shifts):
            demand[f"d{d}_s{s}"] = 1
            
    # Preferencja: Pracownik 0 prosi o wolne 10 i 11 listopada
    penalties = {}
    for d_idx, d_val in enumerate(cal_ctx.dates):
        if d_val in (date(2026, 11, 10), date(2026, 11, 11)):
            for s in range(num_shifts):
                penalties[f"e0_d{d_idx}_s{s}"] = 15

    request = ScheduleRequest(
        num_employees=num_employees,
        start_date=start_date,
        end_date=end_date,
        num_shifts=num_shifts,
        shift_hours=8,
        max_work_hours=kp_norm,
        custom_holidays=custom_holidays,
        closed_days=closed_days,
        auto_calculate_work_norm=True,
        demand=demand,
        shift_penalties=penalties,
        time_limit_seconds=20.0
    )
    
    print("[1/2] Uruchamiam CP-SAT (Google OR-Tools)...")
    cp_res = solve_with_cpsat(request)
    print(f"> Status: {cp_res.status}, Czas: {cp_res.solve_time_seconds:.4f} s, Cel: {cp_res.objective_value}")
    if cp_res.status in ("OPTIMAL", "FEASIBLE"):
        print_calendar_schedule(cp_res.schedule, num_employees, cal_ctx, cp_res.assigned_hours)

    print("[2/2] Uruchamiam SCIP (PySCIPOpt MILP)...")
    scip_res = solve_with_scip(request)
    print(f"> Status: {scip_res.status}, Czas: {scip_res.solve_time_seconds:.4f} s, Cel: {scip_res.objective_value}")
    if scip_res.status in ("OPTIMAL", "FEASIBLE"):
        print_calendar_schedule(scip_res.schedule, num_employees, cal_ctx, scip_res.assigned_hours)

if __name__ == "__main__":
    main()