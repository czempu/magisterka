import sys
import subprocess
import time
import webbrowser
from pathlib import Path

def main():
    root_dir = Path(__file__).resolve().parent
    python_exe = root_dir / ".venv" / "Scripts" / "python.exe"
    rcp_client_exe = root_dir / "rcp_client" / "rcp_client.exe"

    print("=" * 66)
    print("   URUCHAMIANIE SYSTEMU HARMONOGRAMOWANIA PRACY (RCP)")
    print("=" * 66)
    print("-> 1. Silnik optymalizacji (Python FastAPI Backend - port 8000)...")
    backend_proc = subprocess.Popen(
        [str(python_exe), "-m", "uvicorn", "api.main:app", "--host", "0.0.0.0", "--port", "8000"],
        cwd=str(root_dir)
    )

    time.sleep(2)

    print("-> 2. Dashboard i Klient Webowy (Go rcp_client - port 8080)...")
    frontend_proc = subprocess.Popen(
        [str(rcp_client_exe)],
        cwd=str(root_dir / "rcp_client")
    )

    print("\n" + "=" * 66)
    print("   [OK] Wszystkie serwisy dzialaja:")
    print("   * Frontend Dashboard (Go):  http://localhost:8080")
    print("   * Backend Docs (FastAPI):   http://localhost:8000/docs")
    print("=" * 66)
    print("\nNacisnij Ctrl+C, aby zakonczyc prace obu serwisow.\n")

    time.sleep(1)
    try:
        webbrowser.open("http://localhost:8080")
    except Exception:
        pass

    try:
        backend_proc.wait()
        frontend_proc.wait()
    except KeyboardInterrupt:
        print("\nZatrzymywanie serwisow...")
        backend_proc.terminate()
        frontend_proc.terminate()
        backend_proc.wait()
        frontend_proc.wait()
        print("Gotowe.")

if __name__ == "__main__":
    main()
