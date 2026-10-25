import os
from datetime import date
from core.models import ScheduleRequest, EmployeeProfile
from solvers.cp_solver import solve_with_cpsat
from solvers.scip_solver import solve_with_scip
from core.calendar_utils import CalendarContext
from core.excel_exporter import export_schedule_to_excel

def run_demo():
    print("==================================================================")
    print("   DEMO: KWALIFIKACJE PRACOWNIKÓW (SKILLS) & EKSPORT DO EXCELA")
    print("==================================================================\n")
    
    start_date = date(2026, 11, 2)
    end_date = date(2026, 11, 15) # 14 dni
    
    cal_ctx = CalendarContext(start_date=start_date, end_date=end_date)
    norm = cal_ctx.calculate_kp_work_norm_hours() # 72h
    
    # Zespół pracowników ze zróżnicowanymi kwalifikacjami:
    # 3 zmiany dziennie x 14 dni = 42 zmiany x 8h = 336 roboczogodzin
    # Zapotrzebowanie: 1 kierownik na każdej ze zmian -> min. 42 zmiany kierownicze
    employees = [
        EmployeeProfile(id=0, name="Jan Kowalski (Kierownik)", contract_fte=1.0, skills=["kierownik", "pierwsza_pomoc"]),
        EmployeeProfile(id=1, name="Anna Nowak (Kierownik)", contract_fte=1.0, skills=["kierownik"]),
        EmployeeProfile(id=2, name="Piotr Wiśniewski (Kierownik)", contract_fte=1.0, skills=["kierownik", "udt"]),
        EmployeeProfile(id=3, name="Tomasz Wójcik (Kierownik)", contract_fte=1.0, skills=["kierownik"]),
        EmployeeProfile(id=4, name="Katarzyna Kamińska (Med)", contract_fte=1.0, skills=["pierwsza_pomoc"]),
        EmployeeProfile(id=5, name="Marek Lewandowski (Med)", contract_fte=1.0, skills=["pierwsza_pomoc", "udt"]),
        EmployeeProfile(id=6, name="Michał Zieliński (Med)", contract_fte=1.0, skills=["pierwsza_pomoc"]),
        EmployeeProfile(id=7, name="Agnieszka Szymańska (UDT)", contract_fte=1.0, skills=["udt"]),
        EmployeeProfile(id=8, name="Paweł Dąbrowski", contract_fte=1.0, skills=[]),
        EmployeeProfile(id=9, name="Ewa Kozłowska (0.5 FTE)", contract_fte=0.5, skills=[]),
    ]

    # Zapotrzebowanie: 1 osoba na każdą zmianę (lub zmienne)
    demand = {}
    for d in range(cal_ctx.num_days):
        for s in range(3):
            demand[f"d{d}_s{s}"] = 1 # 1 osoba na zmianę (42 zmiany łącznie)

    # Zapotrzebowanie na kwalifikacje (Skills Demand):
    # Na zmianie Rano (s=0) i Popołudnie (s=1) wymagany Kierownik
    # Na zmianie Nocnej (s=2) wymagana Pierwsza Pomoc
    skill_demand = {
        "s0_k_kierownik": 1,        # Na rano zawsze kierownik
        "s1_k_kierownik": 1,        # Na popołudnie zawsze kierownik
        "s2_k_pierwsza_pomoc": 1,   # Na nocki zawsze przeszkolony z pierwszej pomocy
    }

    req = ScheduleRequest(
        employees=employees,
        start_date=start_date,
        end_date=end_date,
        auto_calculate_work_norm=True,
        demand=demand,
        skill_demand=skill_demand,
        time_limit_seconds=15.0
    )

    print("1. Rozwiązywanie problemu za pomocą solvera CP-SAT...")
    cp_res = solve_with_cpsat(req)
    print(f"   > Status: {cp_res.status}, Czas: {cp_res.solve_time_seconds:.4f}s")
    
    # Weryfikacja kwalifikacji na zmianach
    print("\n2. Weryfikacja obsadzenia ról i kwalifikacji na zmianach:")
    skills_ok = True
    for d in range(cal_ctx.num_days):
        for s in range(3):
            working_emps = [e for e in employees if cp_res.schedule.get(f"emp_{e.id}_day_{d}") == s]
            if s in (0, 1): # Rano i Popołudnie wymaga kierownika
                if not any("kierownik" in e.skills for e in working_emps):
                    skills_ok = False
                    print(f"   [!] Dzień {d}, Zmiana {s}: Brak Kierownika! Pracownicy: {[e.name for e in working_emps]}")
            elif s == 2: # Noc wymaga pierwszej pomocy
                if not any("pierwsza_pomoc" in e.skills for e in working_emps):
                    skills_ok = False
                    print(f"   [!] Dzień {d}, Zmiana {s}: Brak Pierwszej Pomocy! Pracownicy: {[e.name for e in working_emps]}")

    if skills_ok:
        print("   [SUKCES] Wszelkie wymagania ról i kwalifikacji zostały spełnione w 100%!")

    # 3. Eksport do pliku Excel
    output_excel_path = "harmonogram_listopad_2026.xlsx"
    print(f"\n3. Generowanie arkusza kalkulacyjnego: {output_excel_path}...")
    with open(output_excel_path, "wb") as f:
        export_schedule_to_excel(cp_res, req, f)
        
    print(f"   [OK] Plik Excel zapisany pomyślnie! Rozmiar: {os.path.getsize(output_excel_path)} bajtów.")

if __name__ == "__main__":
    run_demo()
