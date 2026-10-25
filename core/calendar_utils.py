from datetime import date, timedelta
from typing import List, Set, Tuple

def get_easter_date(year: int) -> date:
    """
    Algorytm Meeusa/Jonesa/Butchera do wyznaczania daty Wielkanocy w kalendarzu gregoriańskim.
    """
    a = year % 19
    b = year // 100
    c = year % 100
    d = b // 4
    e = b % 4
    f = (b + 8) // 25
    g = (b - f + 1) // 3
    h = (19 * a + b - d - g + 15) % 30
    i = c // 4
    k = c % 4
    l = (32 + 2 * e + 2 * i - h - k) % 7
    m = (a + 11 * h + 22 * l) // 451
    month = (h + l - 7 * m + 114) // 31
    day = ((h + l - 7 * m + 114) % 31) + 1
    return date(year, month, day)

def get_polish_public_holidays(year: int) -> Set[date]:
    """
    Zwraca zbiór ustawowo wolnych od pracy dni w Polsce dla danego roku.
    """
    holidays = {
        date(year, 1, 1),   # Nowy Rok
        date(year, 1, 6),   # Święto Trzech Króli
        date(year, 5, 1),   # Święto Pracy
        date(year, 5, 3),   # Święto Konstytucji 3 Maja
        date(year, 8, 15),  # Wniebowzięcie NMP / Święto Wojska Polskiego
        date(year, 11, 1),  # Wszystkich Świętych
        date(year, 11, 11), # Święto Niepodległości
        date(year, 12, 25), # Boże Narodzenie (I dzień)
        date(year, 12, 26), # Boże Narodzenie (II dzień)
    }
    
    # Święta ruchome
    easter = get_easter_date(year)
    easter_monday = easter + timedelta(days=1)
    corpus_christi = easter + timedelta(days=60) # Boże Ciało (czwartek)
    
    holidays.add(easter)
    holidays.add(easter_monday)
    holidays.add(corpus_christi)
    
    return holidays

class CalendarContext:
    def __init__(
        self,
        start_date: date,
        end_date: date,
        custom_holidays: List[date] = None,
        closed_days: List[date] = None,
        shift_hours: int = 8
    ):
        if end_date < start_date:
            raise ValueError(f"end_date ({end_date}) nie może być wcześniejsze niż start_date ({start_date})")

        self.start_date = start_date
        self.end_date = end_date
        self.shift_hours = shift_hours

        # Wszystkie dni w zadanym przedziale
        self.dates: List[date] = []
        curr = start_date
        while curr <= end_date:
            self.dates.append(curr)
            curr += timedelta(days=1)

        self.num_days = len(self.dates)

        # Ustalanie lat występujących w zakresie
        years = {d.year for d in self.dates}
        public_holidays: Set[date] = set()
        for y in years:
            public_holidays.update(get_polish_public_holidays(y))

        # Łączenie świąt
        self.holidays: Set[date] = {d for d in self.dates if d in public_holidays}
        if custom_holidays:
            self.holidays.update(d for d in custom_holidays if start_date <= d <= end_date)

        # Dni zamknięte zakładu
        self.closed_days: Set[date] = set()
        if closed_days:
            self.closed_days.update(d for d in closed_days if start_date <= d <= end_date)

    def is_holiday(self, day_index: int) -> bool:
        return self.dates[day_index] in self.holidays

    def is_closed(self, day_index: int) -> bool:
        return self.dates[day_index] in self.closed_days

    def is_sunday(self, day_index: int) -> bool:
        return self.dates[day_index].weekday() == 6

    def is_saturday(self, day_index: int) -> bool:
        return self.dates[day_index].weekday() == 5

    def get_date_str(self, day_index: int) -> str:
        return self.dates[day_index].strftime("%Y-%m-%d")

    def get_weekday_pl(self, day_index: int) -> str:
        pl_days = ["Pn", "Wt", "Śr", "Cz", "Pt", "Sb", "Nd"]
        return pl_days[self.dates[day_index].weekday()]

    def calculate_kp_work_norm_hours(self) -> int:
        """
        Oblicza wymiar czasu pracy zgodnie z art. 130 Kodeksu Pracy:
        - 40 godzin mnożone przez liczbę pełnych tygodni (lub dni roboczych pon-pt * 8h)
        - każde święto przypadające w okresie rozliczeniowym i przypadające w innym dniu niż niedziela obniża wymiar o 8h.
        """
        standard_work_days = 0
        holidays_reducing_norm = 0

        for d in self.dates:
            weekday = d.weekday() # 0 = poniedziałek, ..., 4 = piątek, 5 = sobota, 6 = niedziela
            if weekday < 5: # poniedziałek - piątek
                standard_work_days += 1
                if d in self.holidays:
                    holidays_reducing_norm += 1
            elif weekday == 5 and d in self.holidays:
                # Święto w sobotę również obniża wymiar czasu pracy o 8h (art. 130 § 2 KP)!
                holidays_reducing_norm += 1

        total_hours = (standard_work_days * self.shift_hours) - (holidays_reducing_norm * self.shift_hours)
        # Opcjonalnie korygujemy o dni robocze pomniejszone o święta
        return max(0, total_hours)
