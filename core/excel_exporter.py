import io
from datetime import datetime
import openpyxl
from openpyxl.styles import Font, PatternFill, Alignment, Border, Side
from openpyxl.utils import get_column_letter

from core.models import ScheduleResponse, ScheduleRequest

def export_schedule_to_excel(response: ScheduleResponse, request: ScheduleRequest, file_path_or_buffer=None):
    """
    Generuje sformatowany plik Microsoft Excel (.xlsx) zawierający:
    - Nagłówek firmy i parametry okresu,
    - Daty i dni tygodnia z wyróżnieniem weekendów i świąt,
    - Przypisane zmiany robocze z czytelnymi kolorami,
    - Oznaczenia urlopów [URLOP] oraz dni wolnych [-],
    - Kolumny z kwalifikacjami, wymiarem etatu (FTE) oraz sumą godzin,
    - Legendę i statystyki wykonania solvera.
    """
    wb = openpyxl.Workbook()
    ws = wb.active
    ws.title = "Grafik Pracy"

    # Style
    font_title = Font(name="Calibri", size=14, bold=True, color="1F497D")
    font_bold = Font(name="Calibri", size=10, bold=True)
    font_header = Font(name="Calibri", size=10, bold=True, color="FFFFFF")
    font_regular = Font(name="Calibri", size=10)
    font_small = Font(name="Calibri", size=9, italic=True)

    fill_header = PatternFill(start_color="1F497D", end_color="1F497D", fill_type="solid")
    fill_weekend = PatternFill(start_color="DCE6F1", end_color="DCE6F1", fill_type="solid")
    fill_holiday = PatternFill(start_color="FDE9D9", end_color="FDE9D9", fill_type="solid") # pomarańczowy pastel
    fill_closed = PatternFill(start_color="F2DCDB", end_color="F2DCDB", fill_type="solid") # czerwony pastel
    fill_vacation = PatternFill(start_color="EBF1DE", end_color="EBF1DE", fill_type="solid") # zielony pastel
    
    # Kolory zmian
    fill_shift_morning = PatternFill(start_color="FFF2CC", end_color="FFF2CC", fill_type="solid")
    fill_shift_afternoon = PatternFill(start_color="D9E1F2", end_color="D9E1F2", fill_type="solid")
    fill_shift_night = PatternFill(start_color="D0CECE", end_color="D0CECE", fill_type="solid")
    
    thin_border = Border(
        left=Side(style='thin', color="D9D9D9"),
        right=Side(style='thin', color="D9D9D9"),
        top=Side(style='thin', color="D9D9D9"),
        bottom=Side(style='thin', color="D9D9D9")
    )
    header_border = Border(
        left=Side(style='thin', color="FFFFFF"),
        right=Side(style='thin', color="FFFFFF"),
        top=Side(style='medium', color="1F497D"),
        bottom=Side(style='medium', color="1F497D")
    )

    # 1. Tytuł i metadane
    ws.merge_cells("A1:G1")
    ws["A1"] = "HARMONOGRAM PRACY (REJESTRACJA CZASU PRACY)"
    ws["A1"].font = font_title
    
    date_range_str = f"{response.calendar_dates[0]} do {response.calendar_dates[-1]}" if response.calendar_dates else "Bieżący okres"
    ws["A2"] = f"Okres: {date_range_str} | Solver: {response.solver_name} ({response.status}) | Czas optymalizacji: {response.solve_time_seconds:.4f}s"
    ws["A2"].font = font_small

    start_row = 4
    
    # 2. Nagłówki kolumn
    ws.cell(row=start_row, column=1, value="ID").fill = fill_header
    ws.cell(row=start_row, column=1).font = font_header
    ws.cell(row=start_row, column=1).alignment = Alignment(horizontal="center", vertical="center")

    ws.cell(row=start_row, column=2, value="Pracownik").fill = fill_header
    ws.cell(row=start_row, column=2).font = font_header
    ws.cell(row=start_row, column=2).alignment = Alignment(horizontal="left", vertical="center")

    ws.cell(row=start_row, column=3, value="Kwalifikacje").fill = fill_header
    ws.cell(row=start_row, column=3).font = font_header
    ws.cell(row=start_row, column=3).alignment = Alignment(horizontal="left", vertical="center")

    ws.cell(row=start_row, column=4, value="Etat (FTE)").fill = fill_header
    ws.cell(row=start_row, column=4).font = font_header
    ws.cell(row=start_row, column=4).alignment = Alignment(horizontal="center", vertical="center")

    num_days = len(response.calendar_dates)
    col_offset = 5 # Pierwszy dzień od kolumny 5 (E)

    # Dni tygodnia i numer dnia
    pl_weekdays = ["Pn", "Wt", "Śr", "Cz", "Pt", "Sb", "Nd"]
    for d_idx, dt_str in enumerate(response.calendar_dates):
        col = col_offset + d_idx
        dt_obj = datetime.strptime(dt_str, "%Y-%m-%d").date()
        wd_str = pl_weekdays[dt_obj.weekday()]
        
        # Wiersz 4: dzień tygodnia + dzień miesiąca
        cell_val = f"{wd_str}\n{dt_obj.day}"
        cell = ws.cell(row=start_row, column=col, value=cell_val)
        cell.font = font_header
        cell.fill = fill_header
        cell.alignment = Alignment(horizontal="center", vertical="center", wrap_text=True)
        cell.border = header_border

    # Kolumny końcowe: Suma godzin, Norma KP, Bilans
    col_sum = col_offset + num_days
    col_norm = col_sum + 1
    col_diff = col_sum + 2

    for c, title in [(col_sum, "Godziny"), (col_norm, "Norma KP"), (col_diff, "Bilans")]:
        cell = ws.cell(row=start_row, column=c, value=title)
        cell.font = font_header
        cell.fill = fill_header
        cell.alignment = Alignment(horizontal="center", vertical="center")
        cell.border = header_border

    ws.row_dimensions[start_row].height = 28

    # 3. Wypełnienie wierszy pracowników
    current_row = start_row + 1
    for emp in request.employees:
        ws.cell(row=current_row, column=1, value=emp.id).alignment = Alignment(horizontal="center")
        ws.cell(row=current_row, column=2, value=emp.name or f"Pracownik {emp.id}")
        ws.cell(row=current_row, column=3, value=", ".join(emp.skills) if emp.skills else "-")
        ws.cell(row=current_row, column=4, value=f"{emp.contract_fte:.2f}").alignment = Alignment(horizontal="center")

        for c in range(1, 5):
            ws.cell(row=current_row, column=c).font = font_regular
            ws.cell(row=current_row, column=c).border = thin_border

        for d_idx, dt_str in enumerate(response.calendar_dates):
            col = col_offset + d_idx
            dt_obj = datetime.strptime(dt_str, "%Y-%m-%d").date()
            cell = ws.cell(row=current_row, column=col)
            cell.border = thin_border
            cell.alignment = Alignment(horizontal="center", vertical="center")

            # Sprawdzenie czy dzień jest zamknięty lub urlop
            shift_id = response.schedule.get(f"emp_{emp.id}_day_{d_idx}")
            is_vacation = emp.vacation_days and dt_obj in emp.vacation_days
            is_closed = dt_str in response.closed_days
            is_holiday = dt_str in response.holidays
            is_weekend = dt_obj.weekday() in (5, 6)

            if is_closed:
                cell.value = "[ZAMK]"
                cell.fill = fill_closed
                cell.font = font_small
            elif is_vacation:
                cell.value = "[URLOP]"
                cell.fill = fill_vacation
                cell.font = font_bold
            elif shift_id is not None:
                # Oznaczenie zmiany (0=Rano, 1=Popołudnie, 2=Noc)
                cell.value = str(shift_id)
                cell.font = font_bold
                if shift_id == 0:
                    cell.fill = fill_shift_morning
                elif shift_id == 1:
                    cell.fill = fill_shift_afternoon
                else:
                    cell.fill = fill_shift_night
            else:
                cell.value = "-"
                cell.font = font_small
                if is_holiday:
                    cell.fill = fill_holiday
                elif is_weekend:
                    cell.fill = fill_weekend

        # Suma i bilans
        assigned_h = response.assigned_hours.get(f"emp_{emp.id}", 0)
        base_norm = response.calculated_work_norm_hours or 0
        emp_norm = int(base_norm * emp.contract_fte)
        diff = assigned_h - emp_norm

        ws.cell(row=current_row, column=col_sum, value=f"{assigned_h}h").alignment = Alignment(horizontal="center")
        ws.cell(row=current_row, column=col_norm, value=f"{emp_norm}h").alignment = Alignment(horizontal="center")
        
        diff_cell = ws.cell(row=current_row, column=col_diff, value=f"{diff:+d}h")
        diff_cell.alignment = Alignment(horizontal="center")
        diff_cell.font = font_bold

        for c in (col_sum, col_norm, col_diff):
            ws.cell(row=current_row, column=c).border = thin_border
            ws.cell(row=current_row, column=c).font = font_bold

        current_row += 1

    # 4. Dostosowanie szerokości kolumn
    ws.column_dimensions["A"].width = 6
    ws.column_dimensions["B"].width = 24
    ws.column_dimensions["C"].width = 20
    ws.column_dimensions["D"].width = 12
    for d in range(num_days):
        col_letter = get_column_letter(col_offset + d)
        ws.column_dimensions[col_letter].width = 6
    ws.column_dimensions[get_column_letter(col_sum)].width = 11
    ws.column_dimensions[get_column_letter(col_norm)].width = 11
    ws.column_dimensions[get_column_letter(col_diff)].width = 10

    # 5. Legenda na dole
    legend_row = current_row + 2
    ws.cell(row=legend_row, column=2, value="LEGENDA ZMIAN I OZNACZEŃ:").font = font_bold
    
    ws.cell(row=legend_row + 1, column=2, value="0").fill = fill_shift_morning
    ws.cell(row=legend_row + 1, column=3, value="Zmiana Rano (06:00 - 14:00)").font = font_regular

    ws.cell(row=legend_row + 2, column=2, value="1").fill = fill_shift_afternoon
    ws.cell(row=legend_row + 2, column=3, value="Zmiana Popołudnie (14:00 - 22:00)").font = font_regular

    ws.cell(row=legend_row + 3, column=2, value="2").fill = fill_shift_night
    ws.cell(row=legend_row + 3, column=3, value="Zmiana Noc (22:00 - 06:00)").font = font_regular

    ws.cell(row=legend_row + 4, column=2, value="[URLOP]").fill = fill_vacation
    ws.cell(row=legend_row + 4, column=3, value="Urlop wypoczynkowy pracownika").font = font_regular

    ws.cell(row=legend_row + 5, column=2, value="[ZAMK]").fill = fill_closed
    ws.cell(row=legend_row + 5, column=3, value="Dzień ustawowo/decyzją zamknięty").font = font_regular

    # Zapis do bufora lub pliku
    if file_path_or_buffer:
        wb.save(file_path_or_buffer)
        return file_path_or_buffer
    else:
        buffer = io.BytesIO()
        wb.save(buffer)
        buffer.seek(0)
        return buffer
