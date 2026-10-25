import uuid
import time
from datetime import datetime, timezone
import threading
from typing import Dict, Optional
import urllib.request
import json
import logging

from core.models import ScheduleRequest, ScheduleResponse, TaskStatusResponse
from solvers.cp_solver import solve_with_cpsat
from solvers.scip_solver import solve_with_scip

logger = logging.getLogger("scheduler_api")

class TaskManager:
    def __init__(self):
        self._tasks: Dict[str, dict] = {}
        self._lock = threading.Lock()

    def create_task(self, request: ScheduleRequest) -> str:
        task_id = str(uuid.uuid4())
        now = datetime.now(timezone.utc)
        with self._lock:
            self._tasks[task_id] = {
                "task_id": task_id,
                "status": "PENDING",
                "solver_choice": request.solver_choice,
                "created_at": now,
                "started_at": None,
                "completed_at": None,
                "execution_time_seconds": None,
                "result": None,
                "error_message": None,
                "request": request
            }
        return task_id

    def get_task_status(self, task_id: str) -> Optional[TaskStatusResponse]:
        with self._lock:
            task = self._tasks.get(task_id)
            if not task:
                return None
            return TaskStatusResponse(
                task_id=task["task_id"],
                status=task["status"],
                solver_choice=task["solver_choice"],
                created_at=task["created_at"],
                started_at=task["started_at"],
                completed_at=task["completed_at"],
                execution_time_seconds=task["execution_time_seconds"],
                result=task["result"],
                error_message=task["error_message"]
            )

    def execute_task_background(self, task_id: str):
        with self._lock:
            task = self._tasks.get(task_id)
            if not task:
                return
            task["status"] = "RUNNING"
            task["started_at"] = datetime.now(timezone.utc)
            request: ScheduleRequest = task["request"]

        start_time = time.perf_counter()
        try:
            if request.solver_choice == "scip":
                result = solve_with_scip(request)
            else:
                result = solve_with_cpsat(request)
                
            execution_time = time.perf_counter() - start_time
            now = datetime.now(timezone.utc)
            
            with self._lock:
                task["status"] = "COMPLETED"
                task["completed_at"] = now
                task["execution_time_seconds"] = execution_time
                task["result"] = result
                
            # Obsługa opcjonalnego webhooka
            if request.webhook_url:
                self._send_webhook(request.webhook_url, task_id, result)

        except Exception as e:
            execution_time = time.perf_counter() - start_time
            now = datetime.now(timezone.utc)
            with self._lock:
                task["status"] = "FAILED"
                task["completed_at"] = now
                task["execution_time_seconds"] = execution_time
                task["error_message"] = str(e)
            logger.exception(f"Błąd podczas rozwiązywania zadania {task_id}: {e}")

    def _send_webhook(self, url: str, task_id: str, result: ScheduleResponse):
        try:
            payload = json.dumps({
                "task_id": task_id,
                "status": "COMPLETED",
                "result": result.model_dump()
            }).encode("utf-8")
            req = urllib.request.Request(
                url,
                data=payload,
                headers={"Content-Type": "application/json"},
                method="POST"
            )
            with urllib.request.urlopen(req, timeout=5) as resp:
                logger.info(f"Webhook {url} wysłany ze statusem {resp.status}")
        except Exception as e:
            logger.warning(f"Nie udało się wysłać webhooka na adres {url}: {e}")

# Globalna instancja TaskManagera
task_manager = TaskManager()
