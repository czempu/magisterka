from datetime import date
from core.models import ScheduleRequest
from solvers.cp_solver import solve_with_cpsat
from solvers.scip_solver import solve_with_scip
from core.calendar_utils import CalendarContext
from core.shift_rules import get_twelve_hour_shift_system, get_default_three_shift_system

def print_shift_system_schedule(schedule_dict, num_employees, cal_ctx: CalendarContext, shifts, assigned_hours=None):
    row_weekdays = "      "
    row_dates = "      "
    row_sep = "      "
    for d in range(cal_ctx.num_days):
        wd = cal_ctx.get_weekday_pl(d)
        d_num = cal_ctx.dates[d].day
        is_sun = cal_ctx.is_sunday(d)
        
        tag = "n" if is_sun else ""
        row_weekdays += f"{wd:<2}{tag:<1} "
        row_dates += f"{d_num:>2}  "
        row_sep += "----"
        
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
                # D dla Dnia (0), N dla Nocy (1) w systemie 12h
                symbol = "D" if shift == 0 else "N" if shift == 1 else str(shift)
                row += f" {symbol}  "
            else:
                row += " -  "
        if assigned_hours:
            hours = assigned_hours.get(f"emp_{e}", 0)
            row += f" |  {hours}h"
        print(row)
    print(row_sep + "\n")
    print(f"Legenda: Zmiany = {[f'{s.id}: {s.name}' for s in shifts]}, [n] = Niedziela (Art. 151.10 KP: min 1 wolna na 4 tyg)\n")

def test_12h_shift_system():
    print("==================================================================")
    print("   TEST SYSTEMU RÓWNOWAŻNEGO: 2x12h (Dzień 06-18, Noc 18-06)")
    print("   Odpoczynek dobowy 11h + Art. 151.10 KP (co 4. niedziela wolna)")
    print("==================================================================\n")
    
    # 28 dni (dokładnie 4 tygodnie = 4 niedziele)
    start_date = date(2026, 10, 1)
    end_date = date(2026, 10, 28)
    
    # System 12-godzinny (2 zmiany: Dzień i Noc)
    shifts_12h = get_twelve_hour_shift_system()
    
    cal_ctx = CalendarContext(
        start_date=start_date,
        end_date=end_date,
        shift_hours=12
    )
    
    num_employees = 6
    demand = {}
    for d in range(cal_ctx.num_days):
        demand[f"d{d}_s0"] = 1 # 1 osoba na dzień (12h)
        demand[f"d{d}_s1"] = 1 # 1 osoba na noc (12h)

    req = ScheduleRequest(
        num_employees=num_employees,
        start_date=start_date,
        end_date=end_date,
        shifts=shifts_12h,
        max_work_hours=120, # Limit 120h (10 zmian po 12h w miesiącu)
        max_consecutive_days=3, # w systemie 12h rzadko pracuje się więcej niż 3 dni z rzędu
        require_sunday_rest_every_4_weeks=True, # Art. 151.10 KP
        enforce_forward_rotation=True, # Ergonomia zmianowa
        demand=demand,
        time_limit_seconds=15.0
    )
    
    print("[1/2] Test CP-SAT dla systemu 2x12h...")
    cp_res = solve_with_cpsat(req)
    print(f"> Status: {cp_res.status}, Czas: {cp_res.solve_time_seconds:.4f} s")
    if cp_res.status in ("OPTIMAL", "FEASIBLE"):
        print_shift_system_schedule(cp_res.schedule, num_employees, cal_ctx, shifts_12h, cp_res.assigned_hours)

    print("[2/2] Test SCIP MILP dla systemu 2x12h...")
    scip_res = solve_with_scip(req)
    print(f"> Status: {scip_res.status}, Czas: {scip_res.solve_time_seconds:.4f} s")
    if scip_res.status in ("OPTIMAL", "FEASIBLE"):
        print_shift_system_schedule(scip_res.schedule, num_employees, cal_ctx, shifts_12h, scip_res.assigned_hours)

if __name__ == "__main__":
    test_12h_shift_system()
