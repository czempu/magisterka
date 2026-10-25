import time
from fastapi.testclient import TestClient
from api.main import app

def test_api_workflow():
    client = TestClient(app)
    
    print("1. Test endpointu /health...")
    health_resp = client.get("/health")
    assert health_resp.status_code == 200
    print("   Healthcheck OK:", health_resp.json())

    # Dane testowe: 6 pracowników, 7 dni, 3 zmiany
    payload = {
        "num_employees": 6,
        "num_days": 7,
        "num_shifts": 3,
        "shift_hours": 8,
        "max_work_hours": 40,
        "max_consecutive_days": 5,
        "solver_choice": "cp-sat",
        "time_limit_seconds": 10.0,
        "shift_penalties": {
            "e0_d0_s0": 10
        }
    }

    print("\n2. Test asynchronicznego zlecenia zadania (/api/v1/schedules/submit)...")
    submit_resp = client.post("/api/v1/schedules/submit", json=payload)
    assert submit_resp.status_code == 202
    data = submit_resp.json()
    task_id = data["task_id"]
    print(f"   Zadanie przyjęte: task_id={task_id}, status={data['status']}")

    print("\n3. Odpytywanie o status zadania (polling /api/v1/schedules/tasks/{id})...")
    final_status = None
    for attempt in range(20):
        time.sleep(0.2)
        status_resp = client.get(f"/api/v1/schedules/tasks/{task_id}")
        assert status_resp.status_code == 200
        status_data = status_resp.json()
        status = status_data["status"]
        print(f"   Próba {attempt+1}: status={status}")
        if status in ("COMPLETED", "FAILED"):
            final_status = status_data
            break

    assert final_status is not None
    assert final_status["status"] == "COMPLETED"
    result = final_status["result"]
    print(f"   Zakończono pomyślnie! Solver: {result['solver_name']}, Czas: {result['solve_time_seconds']:.4f} s")
    print(f"   Liczba przypisanych dni-zmian: {len(result['schedule'])}")
    print(f"   Godziny pracowników: {result['assigned_hours']}")

    print("\n4. Test synchronicznego solvera SCIP (/api/v1/schedules/solve)...")
    payload["solver_choice"] = "scip"
    sync_resp = client.post("/api/v1/schedules/solve", json=payload)
    assert sync_resp.status_code == 200
    scip_res = sync_resp.json()
    print(f"   SCIP OK! Status: {scip_res['status']}, Czas: {scip_res['solve_time_seconds']:.4f} s")

if __name__ == "__main__":
    test_api_workflow()
