import random
from typing import Dict, List
from core.models import ScheduleRequest

def generate_benchmark_suite() -> List[Dict]:
    """
    Tworzy zestaw instancji testowych według założeń experiments_plan.md:
    - Small (Micro-biznes): 5-10 pracowników, 7 dni, 2-3 zmiany
    - Medium (MŚP): 20-30 pracowników, 14-30 dni, 3 zmiany, umiarkowane preferencje
    - Large (Szpital / Zakład ciągły): 50-80 pracowników, 30 dni, 3 zmiany, duże nasycenie preferencji
    """
    random.seed(42) # deterministyczne wyniki
    suite = []

    # 1. Instancje SMALL
    for (name, emps, days, shifts, staff_per_shift) in [
        ("Small-1 (5e_7d)", 5, 7, 2, 1),
        ("Small-2 (7e_7d)", 7, 7, 3, 1),
        ("Small-3 (10e_14d)", 10, 14, 3, 1)
    ]:
        demand = {f"d{d}_s{s}": staff_per_shift for d in range(days) for s in range(shifts)}
        # Kilka preferencji urlopowych
        penalties = {}
        for e in range(emps):
            if random.random() < 0.3:
                d = random.randint(0, days - 1)
                for s in range(shifts):
                    penalties[f"e{e}_d{d}_s{s}"] = 10

        req = ScheduleRequest(
            num_employees=emps,
            num_days=days,
            num_shifts=shifts,
            shift_hours=8,
            max_work_hours=int((days / 7) * 40 * 1.2), # lekki bufor
            max_consecutive_days=5,
            demand=demand,
            shift_penalties=penalties,
            time_limit_seconds=30.0
        )
        suite.append({"id": name, "category": "Small", "request": req})

    # 2. Instancje MEDIUM
    for (name, emps, days, shifts, staff_per_shift) in [
        ("Medium-1 (20e_14d)", 20, 14, 3, 2),
        ("Medium-2 (25e_21d)", 25, 21, 3, 3),
        ("Medium-3 (30e_30d)", 30, 30, 3, 3)
    ]:
        demand = {f"d{d}_s{s}": staff_per_shift for d in range(days) for s in range(shifts)}
        penalties = {}
        for e in range(emps):
            # Preferencje urlopowe (po 2 losowe dni na pracownika)
            for _ in range(2):
                d = random.randint(0, days - 1)
                for s in range(shifts):
                    penalties[f"e{e}_d{d}_s{s}"] = 10
            # Niechęć do nocki u części pracowników
            if e % 3 == 0:
                for d in range(days):
                    penalties[f"e{e}_d{d}_s2"] = 5

        req = ScheduleRequest(
            num_employees=emps,
            num_days=days,
            num_shifts=shifts,
            shift_hours=8,
            max_work_hours=int((days / 7) * 40 * 1.2),
            max_consecutive_days=5,
            demand=demand,
            shift_penalties=penalties,
            time_limit_seconds=45.0
        )
        suite.append({"id": name, "category": "Medium", "request": req})

    # 3. Instancje LARGE
    for (name, emps, days, shifts, staff_per_shift) in [
        ("Large-1 (50e_30d)", 50, 30, 3, 5),
        ("Large-2 (75e_30d)", 75, 30, 3, 8)
    ]:
        demand = {f"d{d}_s{s}": staff_per_shift for d in range(days) for s in range(shifts)}
        penalties = {}
        for e in range(emps):
            for _ in range(3):
                d = random.randint(0, days - 1)
                for s in range(shifts):
                    penalties[f"e{e}_d{d}_s{s}"] = 10

        req = ScheduleRequest(
            num_employees=emps,
            num_days=days,
            num_shifts=shifts,
            shift_hours=8,
            max_work_hours=int((days / 7) * 40 * 1.25),
            max_consecutive_days=5,
            demand=demand,
            shift_penalties=penalties,
            time_limit_seconds=60.0
        )
        suite.append({"id": name, "category": "Large", "request": req})

    return suite
