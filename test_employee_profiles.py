from datetime import date
from core.models import ScheduleRequest, EmployeeProfile, EmployeePreference
from solvers.cp_solver import solve_with_cpsat
from solvers.scip_solver import solve_with_scip
from core.calendar_utils import CalendarContext
from core.shift_rules import get_default_three_shift_system

def print_profile_schedule(res, employees, cal_ctx):
    print("\n      " + "".join([f"{cal_ctx.dates[d].day:<2}{cal_ctx.get_weekday_pl(d):<2} " for d in range(cal_ctx.num_days)]) + " | Godziny | Informacje o etacie i ograniczeniach")
    print("      " + "-" * (cal_ctx.num_days * 5) + " | ------- | -------------------------------------")
    
    for emp in employees:
        row = f"{emp.name[:10]:<10}|"
        for d in range(cal_ctx.num_days):
            shift = res.schedule.get(f"emp_{emp.id}_day_{d}")
            if shift is not None:
                row += f"  {shift}  "
            else:
                if emp.vacation_days and cal_ctx.dates[d] in emp.vacation_days:
                    row += " [U] "
                else:
                    row += "  -  "
        hours = res.assigned_hours.get(f"emp_{emp.id}", 0)
        info = f"FTE: {emp.contract_fte}"
        if emp.disallowed_shift_ids:
            info += f", Zakaz zmian: {emp.disallowed_shift_ids}"
        if emp.vacation_days:
            info += f", Urlop: {[vd.strftime('%d.%m') for vd in emp.vacation_days]}"
        print(f"{row} |  {hours:>3}h   | {info}")
        
    print("      " + "-" * (cal_ctx.num_days * 5) + " | ------- | -------------------------------------\n")
    print("Legenda: [U] Urlop wypoczynkowy, Zmiany: 0=Rano (8h), 1=Popołudnie (8h), 2=Noc (8h)\n")

def test_employee_profiles():
    print("==================================================================")
    print("   TEST ZAAWANSOWANYCH PROFILI PRACOWNIKÓW (EMPLOYEE PROFILE)")
    print("   Zróżnicowane etaty (1.0 vs 0.5), urlopy i zakazy nocek (KP)")
    print("==================================================================\n")
    
    start_date = date(2026, 11, 2)
    end_date = date(2026, 11, 15) # 14 dni (2 tygodnie)
    
    cal_ctx = CalendarContext(start_date=start_date, end_date=end_date)
    norm_100 = cal_ctx.calculate_kp_work_norm_hours() # 72h (obejmuje 11 listopada)
    print(f"Zakres: {start_date} do {end_date} (Dni: {cal_ctx.num_days})")
    print(f"Bazowa norma etatu (1.0 FTE): {norm_100}h\n")
    
    # 7 zróżnicowanych pracowników:
    employees = [
        # Pracownik 0: Pełny etat, ale ma opiekę nad małym dzieckiem (art. 178 KP -> zakaz nocek!)
        EmployeeProfile(id=0, name="Kowalski (Opieka)", contract_fte=1.0, disallowed_shift_ids=[2]),
        
        # Pracownik 1: Pół etatu (0.5 FTE) -> max 36h zamiast 72h
        EmployeeProfile(id=1, name="Nowak (0.5 FTE)", contract_fte=0.5),
        
        # Pracownik 2: Pełny etat, zatwierdzony urlop wypoczynkowy od 5 do 8 listopada
        EmployeeProfile(
            id=2, 
            name="Wiśniewski (Urlop)", 
            contract_fte=1.0, 
            vacation_days=[date(2026, 11, 5), date(2026, 11, 6), date(2026, 11, 7)]
        ),
        
        # Pracownik 3: Student, prosi o wolne w piątki i soboty (miękka prośba)
        EmployeeProfile(
            id=3,
            name="Zieliński (Student)",
            contract_fte=1.0,
            preferences=[
                EmployeePreference(pref_date=date(2026, 11, 6), want_free_day=True, weight=10),
                EmployeePreference(pref_date=date(2026, 11, 7), want_free_day=True, weight=10),
            ]
        ),
        
        # Pozostali pracownicy na pełen etat
        EmployeeProfile(id=4, name="Wójcik (1.0)", contract_fte=1.0),
        EmployeeProfile(id=5, name="Kamińska (1.0)", contract_fte=1.0),
        EmployeeProfile(id=6, name="Lewandowski (1.0)", contract_fte=1.0),
    ]

    demand = {}
    for d in range(cal_ctx.num_days):
        demand[f"d{d}_s0"] = 1 # Rano: 1
        demand[f"d{d}_s1"] = 1 # Popołudnie: 1
        demand[f"d{d}_s2"] = 1 # Noc: 1

    req = ScheduleRequest(
        employees=employees,
        start_date=start_date,
        end_date=end_date,
        auto_calculate_work_norm=True,
        demand=demand,
        time_limit_seconds=15.0
    )
    
    print("[1/2] Test CP-SAT (Google OR-Tools)...")
    cp_res = solve_with_cpsat(req)
    print(f"> Status: {cp_res.status}, Czas: {cp_res.solve_time_seconds:.4f} s")
    if cp_res.status in ("OPTIMAL", "FEASIBLE"):
        print_profile_schedule(cp_res, employees, cal_ctx)

    print("[2/2] Test SCIP (PySCIPOpt MILP)...")
    scip_res = solve_with_scip(req)
    print(f"> Status: {scip_res.status}, Czas: {scip_res.solve_time_seconds:.4f} s")
    if scip_res.status in ("OPTIMAL", "FEASIBLE"):
        print_profile_schedule(scip_res, employees, cal_ctx)

if __name__ == "__main__":
    test_employee_profiles()
