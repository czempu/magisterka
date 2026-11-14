#!/usr/bin/env bash
set -euo pipefail

# --- Konfiguracja (możesz nadpisać przez env) ---
GENERAL_INFO="${GENERAL_INFO:-./cmd/main.go}"                 # plik z @title/@BasePath
OUT_DIR="${OUT_DIR:-./docs/v1}"                     # katalog wyjściowy (RELATYWNY!)
PKG_ROOT="${PKG_ROOT:-.}"                                 # katalog główny modułu
PKG1="${PKG1:-./internal/interface/http/auth}"                    # pakiet z handlerami

# Finalna lista katalogów do skanowania (musi zawierać katalog z GENERAL_INFO)
DIRS="${PKG_ROOT},${PKG1}"

echo "→ Sprawdzam środowisko…"

# 1) swag w PATH
if ! command -v swag >/dev/null 2>&1; then
  echo "✗ 'swag' nie znaleziony w PATH. Zainstaluj: go install github.com/swaggo/swag/cmd/swag@latest"
  exit 1
fi

# 2) go.mod
if [ ! -f "go.mod" ]; then
  echo "✗ Uruchom skrypt w katalogu modułu (tam gdzie jest go.mod)."
  exit 1
fi

# 3) plik GENERAL_INFO
if [ ! -f "${GENERAL_INFO}" ]; then
  echo "✗ Nie znaleziono pliku GENERAL_INFO: ${GENERAL_INFO}"
  exit 1
fi

# 4) Upewnij się, że katalogi w --dir istnieją (sprawdzamy rozdzielone przecinkami)
IFS=',' read -r -a _dirs <<< "${DIRS}"
for d in "${_dirs[@]}"; do
  if [ ! -d "${d}" ]; then
    echo "✗ Katalog z --dir nie istnieje: ${d}"
    exit 1
  fi
done

# 5) OUT_DIR musi być relatywny (żeby nie wylądować w /docs)
case "${OUT_DIR}" in
  /*)
    echo "✗ OUT_DIR jest absolutny (${OUT_DIR}). Użyj ścieżki względnej, np. ./internal/docs"
    exit 1
    ;;
esac

echo "→ Czyszczę ${OUT_DIR}"
rm -rf "${OUT_DIR}"

echo "→ Generuję Swaggera…"
echo "   swag init --dir ${DIRS} --generalInfo ${GENERAL_INFO} --output ${OUT_DIR} --parseDependency --parseInternal"
swag init \
  --dir "${DIRS}" \
  --generalInfo "${GENERAL_INFO}" \
  --output "${OUT_DIR}" \
  --parseDependency \
  --parseInternal

# Szybka weryfikacja
if [ -f "${OUT_DIR}/swagger.json" ]; then
  echo "✓ Wygenerowano: ${OUT_DIR}/swagger.json"
else
  echo "⚠️  Nie znaleziono ${OUT_DIR}/swagger.json — sprawdź logi powyżej."
fi

echo "Gotowe. Zaimportuj pakiet docs w kodzie, np.:"
echo "_ \"$(go list -m)/${OUT_DIR#./}\""

