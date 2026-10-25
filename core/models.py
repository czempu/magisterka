from pydantic import BaseModel, Field, model_validator
from typing import Dict, List, Optional, Literal
import datetime as dt
from core.calendar_utils import CalendarContext
from core.shift_rules import ShiftDefinition, get_default_three_shift_system

class EmployeePreference(BaseModel):
    pref_date: Optional[dt.date] = Field(None, description="Konkretna data preferencji (np. 2026-11-06)", example="2026-11-06")
    day_index: Optional[int] = Field(None, description="Indeks dnia w grafiku (alternatywnie do daty, 0-indeksowany)", example=4)
    preferred_shift_id: Optional[int] = Field(None, description="ID preferowanej zmiany (np. 0 dla Rano)", example=0)
    want_free_day: bool = Field(False, description="Czy pracownik wnioskuje o dzień wolny w tym dniu", example=True)
    weight: int = Field(10, description="Waga kary w funkcji celu w przypadku niespełnienia preferencji (1-100)", example=10)

class EmployeeProfile(BaseModel):
    id: int = Field(..., description="Unikalny identyfikator pracownika", example=0)
    name: str = Field("", description="Imię i nazwisko lub etykieta pracownika", example="Jan Kowalski")
    contract_fte: float = Field(1.0, description="Wymiar etatu (FTE): 1.0 = cały etat, 0.75 = 3/4 etatu, 0.5 = pół etatu", example=1.0)
    skills: List[str] = Field(default_factory=list, description="Lista ról/kwalifikacji (np. kierownik, pierwsza_pomoc, udt)", example=["kierownik", "pierwsza_pomoc"])
    disallowed_shift_ids: List[int] = Field(default_factory=list, description="Prawny lub medyczny zakaz pracy na wskazanych zmianach (np. [2] = brak nocek wg art. 178 KP)", example=[2])
    vacation_days: List[dt.date] = Field(default_factory=list, description="Zatwierdzone urlopy wypoczynkowe (bezwzględny zakaz planowania pracy)", example=["2026-11-05", "2026-11-06"])
    preferences: List[EmployeePreference] = Field(default_factory=list, description="Indywidualne preferencje i wnioski pracownika")

class ScheduleRequest(BaseModel):
    num_employees: Optional[int] = Field(None, description="Liczba pracowników (jeśli nie podano listy 'employees')", example=7)
    employees: List[EmployeeProfile] = Field(default_factory=list, description="Szczegółowe profile pracowników ze stanowiskami, urlopami i preferencjami")
    num_days: Optional[int] = Field(None, description="Liczba dni w okresie (wyliczana automatycznie jeśli podano start_date i end_date)", example=14)
    
    # Konfiguracja systemu zmianowego
    shifts: List[ShiftDefinition] = Field(default_factory=get_default_three_shift_system, description="Definicje zmian roboczych z godzinami startu i długością (domyślnie 3x8h: Rano, Popołudnie, Noc)")
    num_shifts: Optional[int] = Field(None, description="Liczba zmian w dobie (wyliczana automatycznie)")
    shift_hours: Optional[int] = Field(None, description="Domyślna długość zmiany w godzinach (jeśli nie podano w 'shifts')")
    
    # Kodeks Pracy i normy
    max_work_hours: Optional[int] = Field(None, description="Maksymalny limit godzin etatu (jeśli None, wyliczany z art. 130 KP)")
    require_sunday_rest_every_4_weeks: bool = Field(True, description="Art. 151.10 Kodeksu Pracy: wymóg co najmniej 1 wolnej niedzieli na 4 tygodnie w ruchu ciągłym")
    enforce_forward_rotation: bool = Field(False, description="Ergonomia zmianowa: karanie w funkcji celu rotacji wstecznej (np. Noc -> Popołudnie)")
    fairness_balance_nights: bool = True
    
    # Obsługa kalendarza i dat
    start_date: Optional[dt.date] = Field(None, description="Data początkowa harmonogramu (YYYY-MM-DD)", example="2026-10-25")
    end_date: Optional[dt.date] = Field(None, description="Data końcowa harmonogramu (YYYY-MM-DD)", example="2026-11-07")
    custom_holidays: List[dt.date] = Field(default_factory=list, description="Dodatkowe święta lub dni wolne firmy (poza kalendarzem państwowym PL)", example=["2026-11-02"])
    closed_days: List[dt.date] = Field(default_factory=list, description="Dni całkowicie zamknięte dla zakładu (zerowa obsada dla wszystkich)", example=["2026-11-01"])
    auto_calculate_work_norm: bool = Field(True, description="Czy automatycznie wyliczyć normę czasu pracy wg art. 130 Kodeksu Pracy")

    # Zapotrzebowanie ogólne oraz na kwalifikacje (Skills)
    demand: Dict[str, int] = Field(default_factory=dict, description="Zapotrzebowanie na liczbę pracowników: klucz 'd{dzień}_s{zmiana}' -> liczba pracowników", example={"d0_s0": 1, "d0_s1": 1, "d0_s2": 1})
    skill_demand: Dict[str, int] = Field(default_factory=dict, description="Wymogi kwalifikacji: klucz 's{zmiana}_k_{rola}' lub 'd{dzień}_s{zmiana}_k_{rola}' -> min. liczba osób", example={"s0_k_kierownik": 1, "s2_k_pierwsza_pomoc": 1})
    
    shift_penalties: Dict[str, int] = Field(default_factory=dict, description="Tradycyjne kary za przypisanie: 'e{pracownik}_d{dzień}_s{zmiana}' -> waga kary")
    max_consecutive_days: int = Field(6, description="Art. 133 Kodeksu Pracy: maksymalna liczba kolejnych dni pracy z rzędu (odpoczynek 35h)", example=5)
    solver_choice: Literal["cp-sat", "scip"] = Field("cp-sat", description="Wybór silnika optymalizacyjnego: 'cp-sat' (Google OR-Tools) lub 'scip' (PySCIPOpt MILP)", example="cp-sat")
    time_limit_seconds: float = Field(30.0, description="Maksymalny limit czasu obliczeń solvera w sekundach", example=15.0)
    webhook_url: Optional[str] = Field(None, description="Opcjonalny adres Webhook (POST) wywoływany asynchronicznie po zakończeniu obliczeń", example="http://rcp-core:8080/api/v1/schedules/webhook")

    @model_validator(mode="after")
    def validate_setup(self):
        # 1. Walidacja pracowników
        if self.employees:
            self.num_employees = len(self.employees)
        elif self.num_employees is None:
            self.num_employees = 5
            self.employees = [
                EmployeeProfile(id=i, name=f"Pracownik_{i}", contract_fte=1.0)
                for i in range(self.num_employees)
            ]
        else:
            self.employees = [
                EmployeeProfile(id=i, name=f"Pracownik_{i}", contract_fte=1.0)
                for i in range(self.num_employees)
            ]

        # 2. Walidacja zmian
        if not self.shifts:
            self.shifts = get_default_three_shift_system()
        self.num_shifts = len(self.shifts)

        # 3. Walidacja dat
        if self.start_date and self.end_date:
            if self.end_date < self.start_date:
                raise ValueError("end_date nie może być wcześniejsze niż start_date")
            calculated_days = (self.end_date - self.start_date).days + 1
            if self.num_days is None or self.num_days != calculated_days:
                self.num_days = calculated_days
        elif self.num_days is None:
            self.num_days = 7

        return self

class ScheduleResponse(BaseModel):
    solver_name: str = Field(..., description="Użyty algorytm optymalizacyjny (CP-SAT lub SCIP)", example="CP-SAT")
    status: str = Field(..., description="Status rozwiązania: OPTIMAL, FEASIBLE, INFEASIBLE lub TIMEOUT", example="OPTIMAL")
    solve_time_seconds: float = Field(..., description="Czas wykonania obliczeń matematycznych w sekundach", example=0.0245)
    objective_value: Optional[float] = Field(None, description="Wartość funkcji celu (suma kar za naruszone preferencje miękkie)", example=0.0)
    schedule: Dict[str, int] = Field(default_factory=dict, description="Wygenerowany grafik: klucz 'emp_{pracownik}_day_{dzień}' -> id_zmiany")
    assigned_hours: Dict[str, int] = Field(default_factory=dict, description="Zsumowana liczba wypracowanych godzin per pracownik")
    employee_names: Dict[str, str] = Field(default_factory=dict, description="Słownik mapujący ID pracownika na jego imię/nazwisko")
    employee_skills: Dict[str, List[str]] = Field(default_factory=dict, description="Słownik kwalifikacji przypisanych do pracowników")
    
    # Metadane systemu zmianowego i kalendarza
    shift_definitions: List[ShiftDefinition] = Field(default_factory=list, description="Definicje zmian roboczych użyte w optymalizacji")
    calendar_dates: List[str] = Field(default_factory=list, description="Lista dat (YYYY-MM-DD) odpowiadająca indeksom dni grafiku")
    calculated_work_norm_hours: Optional[int] = Field(None, description="Wyliczona norma etatu wg art. 130 Kodeksu Pracy dla tego okresu", example=72)
    holidays: List[str] = Field(default_factory=list, description="Wykryte święta państwowe i firmowe w zadanym okresie")
    closed_days: List[str] = Field(default_factory=list, description="Dni całkowicie wyłączone z pracy (zakład nieczynny)")

class TaskSubmissionResponse(BaseModel):
    task_id: str = Field(..., description="Unikalny identyfikator zadania UUID do odpytywania o status (polling)", example="d2fae3a7-b8cd-4ce6-a5fc-9d6250434bd8")
    status: Literal["PENDING", "RUNNING", "COMPLETED", "FAILED"] = Field("PENDING", description="Aktualny stan przetwarzania", example="PENDING")
    created_at: dt.datetime = Field(..., description="Znacznik czasu utworzenia zadania w UTC")
    message: str = Field(..., description="Komunikat informacyjny dla klienta", example="Zadanie optymalizacji przyjęte pomyślnie. Solver: cp-sat")

class TaskStatusResponse(BaseModel):
    task_id: str = Field(..., description="Identyfikator zadania UUID", example="d2fae3a7-b8cd-4ce6-a5fc-9d6250434bd8")
    status: Literal["PENDING", "RUNNING", "COMPLETED", "FAILED"] = Field(..., description="Aktualny status realizacji zadania w tle", example="COMPLETED")
    solver_choice: str = Field(..., description="Wybrany solver (cp-sat lub scip)", example="cp-sat")
    created_at: dt.datetime = Field(..., description="Czas przyjęcia zlecenia")
    started_at: Optional[dt.datetime] = Field(None, description="Czas rozpoczęcia obliczeń solvera")
    completed_at: Optional[dt.datetime] = Field(None, description="Czas zakończenia obliczeń")
    execution_time_seconds: Optional[float] = Field(None, description="Rzeczywisty łączny czas realizacji zadania", example=0.031)
    result: Optional[ScheduleResponse] = Field(None, description="Kompletny wygenerowany harmonogram pracy (dostępny po statusie COMPLETED)")
    error_message: Optional[str] = Field(None, description="Szczegóły błędu (w przypadku statusu FAILED)")

class BenchmarkResponse(BaseModel):
    cpsat_result: ScheduleResponse = Field(..., description="Wyniki uzyskane z solvera Google OR-Tools CP-SAT")
    scip_result: ScheduleResponse = Field(..., description="Wyniki uzyskane z solvera PySCIPOpt / SCIP MILP")
    cpsat_time_seconds: float = Field(..., description="Czas wykonania CP-SAT w sekundach")
    scip_time_seconds: float = Field(..., description="Czas wykonania SCIP w sekundach")
    speedup_ratio: float = Field(..., description="Stosunek czasu (scip_time / cpsat_time)")
    winner: str = Field(..., description="Nazwa szybszego solvera")
    latex_table_code: str = Field(..., description="Gotowy kod tabeli LaTeX do wklejenia do pracy magisterskiej")

