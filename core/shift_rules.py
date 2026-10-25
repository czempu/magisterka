from typing import List, Tuple
from pydantic import BaseModel, Field

class ShiftDefinition(BaseModel):
    id: int = Field(..., description="Unikalny identyfikator numeryczny zmiany (np. 0, 1, 2)", example=0)
    name: str = Field(..., description="Nazwa zmiany roboczej (np. Rano, Popołudnie, Noc, Dzień 12h)", example="Rano (8h)")
    start_hour: int = Field(..., description="Godzina rozpoczęcia zmiany w formacie 0-23 (np. 6 dla 06:00, 22 dla 22:00)", example=6)
    duration_hours: int = Field(..., description="Czas trwania zmiany w godzinach (np. 8 lub 12)", example=8)

def get_default_three_shift_system() -> List[ShiftDefinition]:
    """
    Standardowy system 3-zmianowy (3x8h):
    - 0: Rano (06:00 - 14:00)
    - 1: Popołudnie (14:00 - 22:00)
    - 2: Noc (22:00 - 06:00)
    """
    return [
        ShiftDefinition(id=0, name="Rano (8h)", start_hour=6, duration_hours=8),
        ShiftDefinition(id=1, name="Popołudnie (8h)", start_hour=14, duration_hours=8),
        ShiftDefinition(id=2, name="Noc (8h)", start_hour=22, duration_hours=8),
    ]

def get_twelve_hour_shift_system() -> List[ShiftDefinition]:
    """
    System równoważny 12-godzinny (2x12h):
    - 0: Dzień (06:00 - 18:00)
    - 1: Noc (18:00 - 06:00)
    """
    return [
        ShiftDefinition(id=0, name="Dzień (12h)", start_hour=6, duration_hours=12),
        ShiftDefinition(id=1, name="Noc (12h)", start_hour=18, duration_hours=12),
    ]

def calculate_forbidden_transitions(shifts: List[ShiftDefinition], min_rest_hours: int = 11) -> List[Tuple[int, int]]:
    """
    Wyznacza pary zmian (s1, s2), które NIE MOGĄ następować dzień po dniu (s1 w dniu d -> s2 w dniu d+1),
    ponieważ czas odpoczynku pomiędzy końcem s1 a początkiem s2 jest mniejszy niż min_rest_hours (domyślnie 11h z art. 132 KP).
    """
    forbidden = []
    for s1 in shifts:
        # Koniec zmiany s1 (względem godziny 00:00 dnia d)
        end_time_s1 = s1.start_hour + s1.duration_hours
        
        for s2 in shifts:
            # Początek zmiany s2 w dniu d+1 (względem godziny 00:00 dnia d, czyli +24h)
            start_time_s2 = 24 + s2.start_hour
            
            rest = start_time_s2 - end_time_s1
            if rest < min_rest_hours:
                forbidden.append((s1.id, s2.id))
                
    return forbidden
