import sys
import logging
from experiments.benchmark_runner import run_benchmarks
from experiments.plot_results import generate_charts

def main():
    print("Rozpoczynanie pełnej procedury eksperymentalnej...")
    run_benchmarks()
    print("\nGenerowanie wykresów analitycznych...")
    generate_charts()
    print("\n[OK] Cała procedura badawcza Fazy 5 zakończona sukcesem!")

if __name__ == "__main__":
    main()
