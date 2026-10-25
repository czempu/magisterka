from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.responses import Response
from datetime import datetime, timezone
import logging

from core.models import (
    ScheduleRequest,
    ScheduleResponse,
    BenchmarkResponse,
    TaskSubmissionResponse,
    TaskStatusResponse
)
from solvers.cp_solver import solve_with_cpsat
from solvers.scip_solver import solve_with_scip
from api.task_manager import task_manager
from core.excel_exporter import export_schedule_to_excel

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("scheduler_api")

tags_metadata = [
    {
        "name": "Harmonogramowanie Asynchroniczne",
        "description": "Zlecanie długotrwałych obliczeń optymalizacyjnych w tle z mechanizmem zadań (task ID), pollingiem oraz webhookami.",
    },
    {
        "name": "Harmonogramowanie Bezpośrednie",
        "description": "Natychmiastowe synchroniczne ułożenie grafiku dla mniejszych instancji lub testów integracyjnych.",
    },
    {
        "name": "Eksport i Raportowanie",
        "description": "Generowanie gotowych, sformatowanych arkuszy kalkulacyjnych Microsoft Excel (.xlsx) z kolorowaniem zmian i podsumowaniem etatu.",
    },
    {
        "name": "System i Diagnostyka",
        "description": "Sprawdzanie stanu zdrowia mikroserwisu dla kontenerów Docker i systemów zewnętrznych.",
    },
]

app = FastAPI(
    title="RCP Scheduler Optimization Engine API",
    description="""
### Inteligentny Silnik Optymalizacji Harmonogramów Pracy (Rejestracja Czasu Pracy)

Mikroserwis optymalizacyjny zrealizowany w ramach pracy magisterskiej, łączący programowanie w logice więzów (**Google OR-Tools CP-SAT**) oraz programowanie mieszane liniowo-całkowitoliczbowe (**PySCIPOpt / SCIP MILP**).

#### Kluczowe możliwości silnika:
* **Zgodność z Kodeksem Pracy**:
  - **Art. 130 KP**: Automatyczne wyliczanie normy czasu pracy w okresie rozliczeniowym (uwzględniające święta pon-sb),
  - **Art. 132 KP**: Gwarancja 11 godzin nieprzerwanego odpoczynku dobowego (dynamiczna weryfikacja godzin startu/końca zmian),
  - **Art. 133 KP**: Odpoczynek tygodniowy 35h (maksymalnie 5–6 kolejnych dni pracy),
  - **Art. 151.10 KP**: Co najmniej jedna wolna niedziela na 4 tygodnie w ruchu ciągłym / pracy zmianowej,
  - **Art. 178 KP**: Wykluczenia pory nocnej dla uprawnionych pracowników (np. opieka nad dzieckiem do 4 lat).
* **Elastyczne Systemy Zmianowe**:
  - Klasyczny 3-zmianowy (3x8h),
  - Równoważny 12-godzinny (2x12h: Dzień / Noc),
  - Dowolne niestandardowe długości i godziny rozpoczęcia zmian.
* **Profile i Preferencje Pracowników**:
  - Zróżnicowane wymiary etatu (FTE: 1.0, 0.75, 0.5),
  - Kwalifikacje i role na zmianie (**Multi-Skill Scheduling**),
  - Zatwierdzone urlopy wypoczynkowe,
  - Miękkie prośby grafikowe optymalizowane w funkcji celu.
* **Architektura Produkcyjna**:
  - Pełna asynchroniczność z identyfikatorami zadań (UUID),
  - Eksport grafiku do Microsoft Excel (.xlsx),
  - Integracja Docker i Docker Compose.
""",
    version="2.0.0",
    openapi_tags=tags_metadata,
    contact={
        "name": "Zespół Optymalizacji Harmonogramów RCP",
        "url": "https://github.com/czempu/magisterka",
    }
)

@app.get(
    "/health",
    tags=["System i Diagnostyka"],
    summary="Sprawdzenie stanu żywotności mikroserwisu",
    description="Zwraca aktualny status silnika, identyfikator usługi oraz znacznik czasu UTC. Używany przez mechanizm HEALTHCHECK w Dockerze."
)
def health_check():
    return {
        "status": "healthy",
        "service": "scheduler_engine",
        "timestamp": datetime.now(timezone.utc).isoformat()
    }

@app.post(
    "/api/v1/schedules/submit",
    response_model=TaskSubmissionResponse,
    status_code=202,
    tags=["Harmonogramowanie Asynchroniczne"],
    summary="Asynchroniczne zlecenie ułożenia harmonogramu pracy",
    description="""
Przyjmuje pełną konfigurację problemu (pracownicy, zapotrzebowanie, system zmianowy, ograniczenia Kodeksu Pracy i kalendarz).

* **Zwraca natychmiast kod HTTP 202 (Accepted)** z unikalnym identyfikatorem `task_id`.
* Obliczenia solvera uruchamiane są w puli zadań w tle (`BackgroundTasks`).
* Stan zadania można następnie odpytywać za pomocą endpointu `GET /api/v1/schedules/tasks/{task_id}`.
* Jeśli podano `webhook_url`, po zakończeniu solver wyśle powiadomienie HTTP POST z wynikiem.
"""
)
def submit_schedule_task(request: ScheduleRequest, background_tasks: BackgroundTasks):
    task_id = task_manager.create_task(request)
    background_tasks.add_task(task_manager.execute_task_background, task_id)
    
    return TaskSubmissionResponse(
        task_id=task_id,
        status="PENDING",
        created_at=datetime.now(timezone.utc),
        message=f"Zadanie optymalizacji przyjęte pomyślnie. Wybrany solver: {request.solver_choice}"
    )

@app.get(
    "/api/v1/schedules/tasks/{task_id}",
    response_model=TaskStatusResponse,
    tags=["Harmonogramowanie Asynchroniczne"],
    summary="Sprawdzenie statusu zadania i odebranie harmonogramu (Polling)",
    description="""
Odpytuje o aktualny stan realizacji zadania optymalizacyjnego:
* **PENDING**: zadanie oczekuje w kolejce,
* **RUNNING**: solver wykonuje obliczenia matematyczne,
* **COMPLETED**: obliczenia zakończone sukcesem – pole `result` zawiera wygenerowany grafik, godziny i metryki,
* **FAILED**: wystąpił błąd podczas obliczeń (szczegóły w `error_message`).
"""
)
def get_task_status(task_id: str):
    status = task_manager.get_task_status(task_id)
    if not status:
        raise HTTPException(status_code=404, detail=f"Nie znaleziono zadania o podanym identyfikatorze: {task_id}")
    return status

@app.post(
    "/api/v1/schedules/solve",
    response_model=ScheduleResponse,
    tags=["Harmonogramowanie Bezpośrednie"],
    summary="Bezpośrednie synchroniczne wygenerowanie harmonogramu",
    description="""
Uruchamia wybrany solver (**CP-SAT** lub **SCIP**) w trybie blokującym (synchronicznym) i zwraca gotowy wynik w odpowiedzi.

Zalecany do:
* Szybkich testów integracyjnych,
* Mniejszych instancji problemu (do kilkunastu pracowników).
"""
)
def solve_synchronously(request: ScheduleRequest):
    try:
        if request.solver_choice == "scip":
            return solve_with_scip(request)
        return solve_with_cpsat(request)
    except Exception as e:
        logger.exception("Błąd podczas synchronicznego rozwiązywania:")
        raise HTTPException(status_code=500, detail=str(e))

@app.post(
    "/api/v1/schedules/export-excel",
    tags=["Eksport i Raportowanie"],
    summary="Optymalizacja i bezpośredni eksport harmonogramu do Microsoft Excel (.xlsx)",
    description="""
Rozwiązuje problem harmonogramowania zgodnie z podanymi ograniczeniami i generuje gotowy, estetycznie sformatowany plik **Microsoft Excel (.xlsx)**.

Zawartość generowanego arkusza:
* Wyróżnienie zmian kolorami pastelowymi,
* Oznaczenie dni weekendowych, świąt państwowych i dni zamkniętych firmy,
* Oznaczenie zatwierdzonych urlopów wypoczynkowych `[URLOP]`,
* Zestawienie kwalifikacji, wymiaru etatu (FTE) oraz bilansu godzin pracownika,
* Pełna legenda kodów zmian na dole arkusza.
"""
)
def export_to_excel_endpoint(request: ScheduleRequest):
    try:
        if request.solver_choice == "scip":
            resp = solve_with_scip(request)
        else:
            resp = solve_with_cpsat(request)
            
        buffer = export_schedule_to_excel(resp, request)
        headers = {
            "Content-Disposition": "attachment; filename=harmonogram_pracy.xlsx"
        }
        return Response(
            content=buffer.getvalue(),
            media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            headers=headers
        )
    except Exception as e:
        logger.exception("Błąd podczas generowania pliku Excel:")
        raise HTTPException(status_code=500, detail=str(e))

@app.post(
    "/api/v1/schedules/benchmark",
    response_model=BenchmarkResponse,
    tags=["Harmonogramowanie Bezpośrednie"],
    summary="Pojedynek solverów: CP-SAT vs SCIP (Benchmark na potrzeby pracy badawczej)",
    description="""
Uruchamia obydwa solvery (**Google OR-Tools CP-SAT** oraz **SCIP MILP**) na identycznym zestawie danych wejściowych,
mierząc dokładny czas obliczeń matematycznych, status rozwiązań oraz wyliczając stosunek przyspieszenia (*speedup*).
Generuje również gotowy kod tabeli **LaTeX** do bezpośredniego wklejenia do pracy magisterskiej.
"""
)
def run_benchmark(request: ScheduleRequest):
    try:
        # 1. Obliczenia CP-SAT
        req_cpsat = request.model_copy(update={"solver_choice": "cp-sat"})
        res_cpsat = solve_with_cpsat(req_cpsat)

        # 2. Obliczenia SCIP
        req_scip = request.model_copy(update={"solver_choice": "scip"})
        res_scip = solve_with_scip(req_scip)

        t_cpsat = max(res_cpsat.solve_time_seconds, 0.0001)
        t_scip = max(res_scip.solve_time_seconds, 0.0001)
        speedup = round(t_scip / t_cpsat, 2)
        winner = "Google OR-Tools CP-SAT" if t_cpsat <= t_scip else "SCIP MILP"

        num_employees = len(request.employees)
        num_days = len(res_cpsat.calendar_dates)

        latex_table = f"""% Tabela wygenerowana automatycznie przez RCP Scheduler Studio
\\begin{{table}}[htbp]
\\centering
\\caption{{Porównanie efektywności algorytmów CP-SAT oraz SCIP MILP (Instancja: {num_employees} pracowników, {num_days} dni)}}
\\label{{tab:benchmark_cpsat_scip}}
\\begin{{tabular}}{{lcccc}}
\\hline
\\textbf{{Algorytm}} & \\textbf{{Status}} & \\textbf{{Czas obliczeń [s]}} & \\textbf{{Funkcja Celu}} & \\textbf{{Przyspieszenie}} \\\\
\\hline
CP-SAT (OR-Tools) & {res_cpsat.status} & {res_cpsat.solve_time_seconds:.4f} s & {res_cpsat.objective_value if res_cpsat.objective_value is not None else '-'} & {speedup:.2f}$\\times$ \\\\
SCIP (PySCIPOpt)  & {res_scip.status}  & {res_scip.solve_time_seconds:.4f} s  & {res_scip.objective_value if res_scip.objective_value is not None else '-'} & 1.00$\\times$ \\\\
\\hline
\\end{{tabular}}
\\end{{table}}"""

        return BenchmarkResponse(
            cpsat_result=res_cpsat,
            scip_result=res_scip,
            cpsat_time_seconds=res_cpsat.solve_time_seconds,
            scip_time_seconds=res_scip.solve_time_seconds,
            speedup_ratio=speedup,
            winner=winner,
            latex_table_code=latex_table
        )
    except Exception as e:
        logger.exception("Błąd podczas testu benchmarkowego:")
        raise HTTPException(status_code=500, detail=str(e))

