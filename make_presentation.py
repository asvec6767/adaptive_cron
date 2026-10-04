# -*- coding: utf-8 -*-
"""
Генератор презентации Adaptive Cron.
Требуется: pip install python-pptx
Результат: adaptive_cron.pptx (16:9, совместим с классическим PowerPoint)
"""

from pptx import Presentation
from pptx.util import Inches, Pt, Emu
from pptx.dml.color import RGBColor
from pptx.enum.text import PP_ALIGN, MSO_ANCHOR
from pptx.enum.shapes import MSO_SHAPE

# ---------- Цвета и шрифты ----------
BG_DARK    = RGBColor(0x1B, 0x26, 0x33)
BG_LIGHT   = RGBColor(0xF5, 0xF7, 0xFA)
ACCENT     = RGBColor(0x2E, 0x86, 0xDE)
ACCENT_2   = RGBColor(0x27, 0xAE, 0x60)
ACCENT_3   = RGBColor(0xE6, 0x7E, 0x22)
TEXT_DARK  = RGBColor(0x1B, 0x26, 0x33)
TEXT_LIGHT = RGBColor(0xFF, 0xFF, 0xFF)
TEXT_MUTED = RGBColor(0x7F, 0x8C, 0x8D)
CODE_BG    = RGBColor(0x1E, 0x1E, 0x1E)
CODE_FG    = RGBColor(0xE6, 0xE6, 0xE6)
CODE_ACCENT= RGBColor(0x9C, 0xDC, 0xFE)

FONT_MAIN = "Calibri"
FONT_CODE = "Consolas"

# ---------- Создание презентации ----------
prs = Presentation()
prs.slide_width  = Inches(13.333)
prs.slide_height = Inches(7.5)
BLANK = prs.slide_layouts[6]

def _set_bg(slide, color):
    fill = slide.background.fill
    fill.solid()
    fill.fore_color.rgb = color

def _add_text(slide, left, top, width, height, text, *,
              size=18, bold=False, color=TEXT_DARK, align=PP_ALIGN.LEFT,
              font=FONT_MAIN, anchor=MSO_ANCHOR.TOP):
    box = slide.shapes.add_textbox(Inches(left), Inches(top),
                                   Inches(width), Inches(height))
    tf = box.text_frame
    tf.word_wrap = True
    tf.vertical_anchor = anchor
    p = tf.paragraphs[0]
    p.alignment = align
    run = p.add_run()
    run.text = text
    run.font.name = font
    run.font.size = Pt(size)
    run.font.bold = bold
    run.font.color.rgb = color
    return box

def _add_accent_bar(slide, left, top, width=0.08, height=1.0, color=ACCENT):
    bar = slide.shapes.add_shape(MSO_SHAPE.RECTANGLE,
                                 Inches(left), Inches(top),
                                 Inches(width), Inches(height))
    bar.fill.solid()
    bar.fill.fore_color.rgb = color
    bar.line.fill.background()
    return bar

# ---------- Типовые макеты ----------

def slide_title(title, subtitle, footer):
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_DARK)
    _add_accent_bar(s, 0.8, 2.6, 0.12, 1.8, ACCENT)
    _add_text(s, 1.2, 2.5, 11, 1.2, title,
              size=48, bold=True, color=TEXT_LIGHT)
    _add_text(s, 1.2, 3.9, 11, 0.8, subtitle,
              size=22, color=ACCENT)
    _add_text(s, 1.2, 6.6, 11, 0.5, footer,
              size=14, color=TEXT_MUTED)
    return s

def slide_section(title, items, *, accent=ACCENT, columns=1):
    """Слайд с заголовком и списком пунктов (текст или (текст, уровень))."""
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, accent)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, title,
              size=34, bold=True, color=TEXT_DARK)

    if columns == 2:
        half = len(items) // 2 + len(items) % 2
        groups = [items[:half], items[half:]]
        for ci, group in enumerate(groups):
            box = s.shapes.add_textbox(
                Inches(0.9 + ci * 6.1), Inches(1.7),
                Inches(5.7), Inches(5.3))
            _fill_bullets(box.text_frame, group, accent)
    else:
        box = s.shapes.add_textbox(
            Inches(0.9), Inches(1.7), Inches(11.5), Inches(5.3))
        _fill_bullets(box.text_frame, items, accent)
    return s

def _fill_bullets(tf, items, accent):
    tf.word_wrap = True
    first = True
    for item in items:
        if isinstance(item, tuple):
            text, level = item
        else:
            text, level = item, 0
        p = tf.paragraphs[0] if first else tf.add_paragraph()
        first = False
        p.level = level
        p.space_after = Pt(8)
        bullet = "● " if level == 0 else "– "
        run = p.add_run()
        run.text = bullet + text
        run.font.name = FONT_MAIN
        run.font.size = Pt(20 - level * 3)
        run.font.color.rgb = TEXT_DARK if level == 0 else TEXT_MUTED
        if level == 0:
            run.font.bold = False

def slide_table(title, headers, rows, *, accent=ACCENT, col_widths=None):
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, accent)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, title,
              size=34, bold=True, color=TEXT_DARK)

    n_rows, n_cols = len(rows) + 1, len(headers)

    # --- Автоподбор высоты строки и шрифта, чтобы таблица влезла в слайд ---
    top_in = 1.8                      # верх таблицы, дюймы
    bottom_margin_in = 0.4            # запас снизу, дюймы
    max_height_in = prs.slide_height.inches - top_in - bottom_margin_in
    row_h_in = min(0.55, max_height_in / n_rows)
    font_size = 16 if row_h_in >= 0.5 else (14 if row_h_in >= 0.42 else 12)
    head_size = font_size

    left, top = Inches(0.9), Inches(top_in)
    width = Inches(11.5)
    height = Inches(row_h_in * n_rows)

    table = s.shapes.add_table(n_rows, n_cols, left, top, width, height).table

    if col_widths:
        total = sum(col_widths)
        for i, w in enumerate(col_widths):
            table.columns[i].width = Emu(int(width * w / total))

    for c, h in enumerate(headers):
        cell = table.cell(0, c)
        cell.text = h
        cell.fill.solid()
        cell.fill.fore_color.rgb = accent
        _style_cell(cell, color=TEXT_LIGHT, bold=True, size=head_size)

    for r, row in enumerate(rows, start=1):
        for c, val in enumerate(row):
            cell = table.cell(r, c)
            cell.text = str(val)
            cell.fill.solid()
            cell.fill.fore_color.rgb = (
                RGBColor(0xFF, 0xFF, 0xFF) if r % 2 else RGBColor(0xEC, 0xF0, 0xF6)
            )
            _style_cell(cell, color=TEXT_DARK, bold=False, size=font_size)
    return s

def _style_cell(cell, *, color, bold, size):
    cell.vertical_anchor = MSO_ANCHOR.MIDDLE
    for p in cell.text_frame.paragraphs:
        p.alignment = PP_ALIGN.LEFT
        for run in p.runs:
            run.font.name = FONT_MAIN
            run.font.size = Pt(size)
            run.font.bold = bold
            run.font.color.rgb = color

def slide_code(title, code, *, accent=ACCENT):
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, accent)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, title,
              size=34, bold=True, color=TEXT_DARK)

    box = s.shapes.add_textbox(Inches(0.9), Inches(1.8),
                               Inches(11.5), Inches(5.2))
    box.fill.solid()
    box.fill.fore_color.rgb = CODE_BG
    box.line.fill.background()
    tf = box.text_frame
    tf.word_wrap = True
    tf.margin_left = Inches(0.3)
    tf.margin_top = Inches(0.25)
    for i, line in enumerate(code.strip("\n").splitlines()):
        p = tf.paragraphs[0] if i == 0 else tf.add_paragraph()
        run = p.add_run()
        run.text = line
        run.font.name = FONT_CODE
        run.font.size = Pt(14)
        run.font.color.rgb = CODE_FG
    return s

def slide_terminal(title, blocks, *, accent=ACCENT_2):
    """
    Слайд-«терминал». blocks = список кортежей (command, output).
    command — строка приглашения, output — многострочный вывод.
    """
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, accent)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, title,
              size=34, bold=True, color=TEXT_DARK)

    box = s.shapes.add_textbox(Inches(0.9), Inches(1.8),
                               Inches(11.5), Inches(5.2))
    box.fill.solid()
    box.fill.fore_color.rgb = CODE_BG
    box.line.fill.background()
    tf = box.text_frame
    tf.word_wrap = True
    tf.margin_left = Inches(0.35)
    tf.margin_top = Inches(0.25)
    tf.margin_right = Inches(0.35)
    tf.margin_bottom = Inches(0.25)

    first_par = True
    for cmd, out in blocks:
        # строка приглашения: $ команда
        p = tf.paragraphs[0] if first_par else tf.add_paragraph()
        first_par = False
        p.space_before = Pt(6)
        run = p.add_run()
        run.text = "$ " + cmd
        run.font.name = FONT_CODE
        run.font.size = Pt(13)
        run.font.bold = True
        run.font.color.rgb = CODE_ACCENT

        # вывод
        if out:
            for line in out.strip("\n").splitlines():
                p = tf.add_paragraph()
                run = p.add_run()
                run.text = line
                run.font.name = FONT_CODE
                run.font.size = Pt(12)
                run.font.color.rgb = CODE_FG
    return s

def slide_architecture():
    """Слайд с компонентами архитектуры в виде карточек."""
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, ACCENT)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, "Архитектура системы",
              size=34, bold=True, color=TEXT_DARK)

    components = [
        ("API Gateway", "8080", "Единая точка входа,\nмаршрутизация", ACCENT),
        ("Task Manager", "8081", "CRUD операций\nс задачами", ACCENT_2),
        ("Scheduler", "8082", "Адаптивное\nпланирование", ACCENT_3),
        ("Metric Agent", "8083", "CPU, память,\nload average", ACCENT),
        ("Worker", "8084", "Выполнение\nкоманд", ACCENT_2),
        ("CLI", "—", "Встроен в единый\nбинарник", ACCENT_3),
    ]
    cols, card_w, card_h = 3, 3.7, 2.0
    gap_x, gap_y = 0.35, 0.35
    start_x, start_y = 0.9, 1.8

    for i, (name, port, desc, color) in enumerate(components):
        r, c = divmod(i, cols)
        left = start_x + c * (card_w + gap_x)
        top  = start_y + r * (card_h + gap_y)

        card = s.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE,
                                  Inches(left), Inches(top),
                                  Inches(card_w), Inches(card_h))
        card.fill.solid()
        card.fill.fore_color.rgb = RGBColor(0xFF, 0xFF, 0xFF)
        card.line.color.rgb = color
        card.line.width = Pt(1.5)

        _add_text(s, left + 0.2, top + 0.15, card_w - 0.4, 0.5,
                  name, size=18, bold=True, color=color)
        _add_text(s, left + 0.2, top + 0.65, card_w - 0.4, 0.4,
                  f"порт {port}", size=13, color=TEXT_MUTED)
        _add_text(s, left + 0.2, top + 1.15, card_w - 0.4, 0.8,
                  desc, size=13, color=TEXT_DARK)
    return s

def _box(s, left, top, w, h, text, *, fill, border, text_color=TEXT_DARK,
         size=13, bold=True):
    shape = s.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE,
                               Inches(left), Inches(top),
                               Inches(w), Inches(h))
    shape.fill.solid()
    shape.fill.fore_color.rgb = fill
    shape.line.color.rgb = border
    shape.line.width = Pt(1.5)
    tf = shape.text_frame
    tf.word_wrap = True
    tf.margin_left = Inches(0.05)
    tf.margin_right = Inches(0.05)
    p = tf.paragraphs[0]
    p.alignment = PP_ALIGN.CENTER
    run = p.add_run()
    run.text = text
    run.font.name = FONT_MAIN
    run.font.size = Pt(size)
    run.font.bold = bold
    run.font.color.rgb = text_color
    return shape

def _arrow(s, x1, y1, x2, y2, *, color=ACCENT, width=Pt(1.75), dashed=False):
    conn = s.shapes.add_connector(
        2,  # MSO_CONNECTOR.STRAIGHT
        Inches(x1), Inches(y1), Inches(x2), Inches(y2),
    )
    conn.line.color.rgb = color
    conn.line.width = width
    if dashed:
        from pptx.enum.dml import MSO_LINE_DASH_STYLE
        conn.line.dash_style = MSO_LINE_DASH_STYLE.DASH
    # стрелка на конце
    from pptx.oxml.ns import qn
    ln = conn.line._get_or_add_ln()
    tail = ln.makeelement(qn('a:tailEnd'), {'type': 'triangle'})
    ln.append(tail)
    return conn

def slide_architecture_diagram():
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, ACCENT)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, "Диаграмма архитектуры",
              size=34, bold=True, color=TEXT_DARK)

    # --- Верхний уровень: клиент ---
    _box(s, 5.4, 1.7, 2.5, 0.7, "Клиент / CLI",
         fill=RGBColor(0xE8, 0xF1, 0xFB), border=ACCENT)

    # --- API Gateway ---
    _box(s, 5.4, 2.9, 2.5, 0.8, "API Gateway :8080\n(REST)",
         fill=RGBColor(0xFF, 0xFF, 0xFF), border=ACCENT, size=12)

    # --- Сервисы (3 в ряд) ---
    services = [
        ("Task Manager :8081", "CRUD задач", ACCENT_2),
        ("Scheduler :8082", "Адаптивное планирование", ACCENT_3),
        ("Metric Agent :8083", "CPU / Mem / Load", ACCENT),
    ]
    box_w, box_h, gap = 3.5, 1.3, 0.35
    start_x = (prs.slide_width.inches - (box_w * 3 + gap * 2)) / 2
    y = 4.15
    centers = []
    for i, (name, sub, color) in enumerate(services):
        left = start_x + i * (box_w + gap)
        _box(s, left, y, box_w, box_h, f"{name}\n{sub}",
             fill=RGBColor(0xFF, 0xFF, 0xFF), border=color, size=12)
        centers.append(left + box_w / 2)

    # --- Worker снизу ---
    _box(s, 5.4, 5.9, 2.5, 0.8, "Worker :8084",
         fill=RGBColor(0xFF, 0xFF, 0xFF), border=ACCENT_2, size=12)

    # --- Инфраструктура (тонкая строка) ---
    _box(s, 0.9, 6.95, 11.5, 0.4,
         "Инфраструктура:   SQLite / PostgreSQL   ·   In-Memory / Redis   ·   Файл / ClickHouse+Grafana   ·   Prometheus",
         fill=RGBColor(0xEC, 0xF0, 0xF6), border=TEXT_MUTED,
         text_color=TEXT_DARK, size=11, bold=False)

    # --- Стрелки: клиент → gateway ---
    _arrow(s, 6.65, 2.4, 6.65, 2.9)

    # --- gateway → сервисы ---
    gy = 3.7
    for cx in centers:
        _arrow(s, 6.65, gy, cx, y)

    # --- scheduler → worker ---
    _arrow(s, centers[1], y + box_h, 6.65, 5.9)

    # --- Worker → инфраструктура ---
    _arrow(s, 6.65, 6.7, 6.65, 6.95, dashed=True, color=TEXT_MUTED)

def slide_algorithm():
    """Слайд с формулой Load Score и шагами алгоритма."""
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, ACCENT_3)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, "Адаптивный алгоритм",
              size=34, bold=True, color=TEXT_DARK)

    # Формула
    formula_box = s.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE,
                                     Inches(0.9), Inches(1.7),
                                     Inches(11.5), Inches(1.0))
    formula_box.fill.solid()
    formula_box.fill.fore_color.rgb = RGBColor(0xFF, 0xF7, 0xE6)
    formula_box.line.color.rgb = ACCENT_3
    _add_text(s, 1.1, 1.9, 11.1, 0.6,
              "Load Score = CPU% × 0.4 + Memory% × 0.3 + (LoadAvg1 / cores) × 0.3",
              size=18, bold=True, color=TEXT_DARK, align=PP_ALIGN.CENTER)

    steps = [
        "1. Задача критическая (is_critical=true) → запуск немедленно",
        "2. CPU > max_cpu_percent → отложить на 30 секунд",
        "3. Memory > max_memory_mb → отложить на 30 секунд",
        "4. Load Score > max_load_score → отложить на score × 60 секунд",
        "5. Все проверки пройдены → запуск через Worker",
    ]
    box = s.shapes.add_textbox(Inches(0.9), Inches(3.0),
                               Inches(11.5), Inches(4.0))
    tf = box.text_frame
    tf.word_wrap = True
    for i, line in enumerate(steps):
        p = tf.paragraphs[0] if i == 0 else tf.add_paragraph()
        p.space_after = Pt(10)
        run = p.add_run()
        run.text = line
        run.font.name = FONT_MAIN
        run.font.size = Pt(18)
        run.font.color.rgb = TEXT_DARK
    return s

def slide_profiles():
    """Слайд сравнения трёх профилей."""
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, ACCENT_2)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, "Три профиля конфигурации",
              size=34, bold=True, color=TEXT_DARK)

    profiles = [
        ("Lite", ACCENT_2, [
            "Транспорт: In-Memory",
            "БД: SQLite",
            "Очередь: In-Memory",
            "Логи: файл",
            "Метрики: встроенные",
            "Для разработки и\nнебольших задач",
        ]),
        ("Standard", ACCENT, [
            "Транспорт: REST",
            "БД: PostgreSQL",
            "Очередь: Redis",
            "Логи: JSON-файл",
            "Метрики: Prometheus",
            "Для production",
        ]),
        ("Enterprise", ACCENT_3, [
            "Транспорт: REST (gRPC/NATS — TODO)",
            "БД: PostgreSQL",
            "Очередь: Redis",
            "Логи: ClickHouse + Grafana",
            "Метрики: Prometheus",
            "Для крупных систем",
        ]),
    ]
    card_w, card_h = 3.7, 4.6
    gap = 0.35
    start_x, start_y = 0.9, 1.8

    for i, (name, color, items) in enumerate(profiles):
        left = start_x + i * (card_w + gap)
        card = s.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE,
                                  Inches(left), Inches(start_y),
                                  Inches(card_w), Inches(card_h))
        card.fill.solid()
        card.fill.fore_color.rgb = RGBColor(0xFF, 0xFF, 0xFF)
        card.line.color.rgb = color
        card.line.width = Pt(2)

        _add_text(s, left + 0.25, start_y + 0.2, card_w - 0.5, 0.6,
                  name, size=24, bold=True, color=color)

        box = s.shapes.add_textbox(Inches(left + 0.25), Inches(start_y + 0.9),
                                   Inches(card_w - 0.5), Inches(card_h - 1.1))
        tf = box.text_frame
        tf.word_wrap = True
        for j, item in enumerate(items):
            p = tf.paragraphs[0] if j == 0 else tf.add_paragraph()
            p.space_after = Pt(6)
            run = p.add_run()
            run.text = "• " + item
            run.font.name = FONT_MAIN
            run.font.size = Pt(13)
            run.font.color.rgb = TEXT_DARK
    return s

def slide_roadmap():
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_LIGHT)
    _add_accent_bar(s, 0.6, 0.55, 0.1, 0.9, ACCENT)
    _add_text(s, 0.9, 0.5, 11.5, 1.0, "Roadmap",
              size=34, bold=True, color=TEXT_DARK)

    versions = [
        ("1.0", "Текущая", ACCENT_2, [
            "Архитектура, Clean Code, профили",
            "CLI-интерфейс",
            "Docker Compose",
            "In-Memory и REST транспорт",
            "SQLite, PostgreSQL, Redis",
        ]),
        ("1.1", "В разработке", ACCENT, [
            "gRPC для внутренней коммуникации",
            "NATS транспорт",
            "ClickHouse логирование",
            "Prometheus + Grafana",
            "Web UI",
        ]),
        ("2.0", "Планы", ACCENT_3, [
            "Горизонтальное масштабирование",
            "Leader Election для Scheduler",
            "ML для предсказания нагрузки",
            "Kubernetes Helm",
            "Service Discovery (etcd/Consul)",
        ]),
    ]
    card_w, card_h = 3.7, 4.6
    gap = 0.35
    start_x, start_y = 0.9, 1.8

    for i, (ver, status, color, items) in enumerate(versions):
        left = start_x + i * (card_w + gap)
        card = s.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE,
                                  Inches(left), Inches(start_y),
                                  Inches(card_w), Inches(card_h))
        card.fill.solid()
        card.fill.fore_color.rgb = RGBColor(0xFF, 0xFF, 0xFF)
        card.line.color.rgb = color
        card.line.width = Pt(2)

        _add_text(s, left + 0.25, start_y + 0.2, card_w - 0.5, 0.5,
                  f"Версия {ver}", size=22, bold=True, color=color)
        _add_text(s, left + 0.25, start_y + 0.75, card_w - 0.5, 0.4,
                  status, size=13, color=TEXT_MUTED)

        box = s.shapes.add_textbox(Inches(left + 0.25), Inches(start_y + 1.25),
                                   Inches(card_w - 0.5), Inches(card_h - 1.5))
        tf = box.text_frame
        tf.word_wrap = True
        for j, item in enumerate(items):
            p = tf.paragraphs[0] if j == 0 else tf.add_paragraph()
            p.space_after = Pt(6)
            run = p.add_run()
            run.text = "• " + item
            run.font.name = FONT_MAIN
            run.font.size = Pt(13)
            run.font.color.rgb = TEXT_DARK
    return s

def slide_final(title, lines):
    s = prs.slides.add_slide(BLANK)
    _set_bg(s, BG_DARK)
    _add_accent_bar(s, 0.8, 2.5, 0.12, 1.6, ACCENT)
    _add_text(s, 1.2, 2.4, 11, 1.0, title,
              size=44, bold=True, color=TEXT_LIGHT)
    box = s.shapes.add_textbox(Inches(1.2), Inches(3.9),
                               Inches(11), Inches(3.0))
    tf = box.text_frame
    tf.word_wrap = True
    for i, line in enumerate(lines):
        p = tf.paragraphs[0] if i == 0 else tf.add_paragraph()
        p.space_after = Pt(10)
        run = p.add_run()
        run.text = line
        run.font.name = FONT_MAIN
        run.font.size = Pt(20)
        run.font.color.rgb = ACCENT
    return s

# ============================================================
#                    СБОРКА ПРЕЗЕНТАЦИИ
# ============================================================

# 1. Титул
slide_title(
    "Adaptive Cron",
    "Адаптивный планировщик задач с микросервисной архитектурой",
    "Версия 1.0 · MIT License · 2026",
)

# 2. Проблема и решение
slide_section(
    "Проблема и решение",
    [
        ("Проблема", 0),
        ("Классический cron не учитывает нагрузку сервера", 1),
        ("Запуск тяжёлых задач в пик перегружает систему", 1),
        ("Нет гибкости: невозможно приоритизировать задачи", 1),
        ("Решение — Adaptive Cron", 0),
        ("Адаптивное планирование на основе метрик CPU, памяти, load average", 1),
        ("Критические задачи выполняются всегда", 1),
        ("Три профиля под разные масштабы: Lite / Standard / Enterprise", 1),
    ],
    accent=ACCENT,
)

# 3. Ключевые возможности
slide_section(
    "Ключевые возможности",
    [
        ("Адаптивное планирование на основе системных метрик", 0),
        ("Критические задачи (выполняются всегда)", 0),
        ("Микросервисная архитектура с Clean Code слоями", 0),
        ("Feature Flags для гибкой настройки", 0),
        ("Три профиля: Lite, Standard, Enterprise", 0),
        ("CLI интерфейс — встроен в единый бинарник", 0),
        ("Горизонтальное масштабирование (в планах 2.0)", 0),
        ("Машинное обучение для предсказания нагрузки (в планах 2.0)", 0),
    ],
    accent=ACCENT_2,
)

# 4. Профили
slide_profiles()

# 5. Архитектура
slide_architecture()

# 5b. Диаграмма архитектуры
slide_architecture_diagram()

# 6. Clean Code слои
slide_section(
    "Clean Code слои",
    [
        ("Каждый сервис в internal/ имеет четыре слоя", 0),
        ("Domain — сущности, ошибки, константы. Без внешних зависимостей", 1),
        ("Repository — интерфейсы доступа к данным", 1),
        ("UseCase — бизнес-логика, координирует слои", 1),
        ("Delivery — обработка запросов (HTTP)", 1),
        ("Зависимости направлены внутрь", 0),
        ("Domain ← Repository ← UseCase ← Delivery", 1),
        ("Общие модели (Task, Metric, Log) — в internal/common/models", 0),
    ],
    accent=ACCENT,
)

# 7. Транспортный слой
slide_table(
    "Транспортный слой",
    ["Транспорт", "Профиль", "Статус"],
    [
        ["In-Memory", "Lite", "Реализован"],
        ["REST", "Standard, Enterprise", "Реализован"],
        ["gRPC", "Enterprise", "TODO"],
        ["NATS", "Enterprise", "TODO"],
    ],
    accent=ACCENT,
)

# 8. Алгоритм
slide_algorithm()

# 9. Feature Flags
slide_section(
    "Feature Flags",
    [
        ("Загружаются из конфига при старте", 0),
        ("Имеют значения по умолчанию для каждого профиля", 0),
        ("Примеры флагов:", 0),
        ("db.sqlite / db.postgres", 1),
        ("queue.memory / queue.redis", 1),
        ("log.file / log.clickhouse", 1),
        ("transport.inmemory / transport.rest / transport.grpc / transport.nats", 1),
        ("metrics.basic / metrics.prometheus", 1),
        ("scheduler.local / scheduler.distributed (TODO)", 1),
    ],
    accent=ACCENT_3,
)

# 10. Технологический стек
slide_table(
    "Технологический стек",
    ["Категория", "Технологии"],
    [
        ["Язык", "Go 1.21+"],
        ["HTTP", "gorilla/mux"],
        ["Конфигурация", "viper (YAML)"],
        ["Логирование", "logrus"],
        ["ORM", "GORM (с возможностью замены на SQL)"],
        ["Cron-парсер", "robfig/cron"],
        ["Тестирование", "testify"],
        ["БД", "SQLite / PostgreSQL"],
        ["Очередь", "In-Memory / Redis"],
        ["Метрики", "встроенные / Prometheus"],
    ],
    accent=ACCENT,
    col_widths=[2, 5],
)

# 11. Запуск
slide_code(
    "Запуск",
    """
# Сборка
make build

# Запуск по профилям
./bin/adaptive-cron --profile lite
./bin/adaptive-cron --profile standard --config configs/standard.yaml
./bin/adaptive-cron --profile enterprise --config configs/enterprise.yaml

# Список профилей
./bin/adaptive-cron --list-profiles

# Включить конкретный feature flag
./bin/adaptive-cron --profile lite --enable metrics.prometheus

# Docker
make docker-lite
make docker-standard
make docker-enterprise
""",
    accent=ACCENT_2,
)

# 11b. Скриншоты работы (терминал)
slide_terminal(
    "Как это выглядит в работе",
    [
        (
            "./bin/adaptive-cron --profile lite",
            """
2026-10-05T10:00:01 INFO  profile=lite transport=in-memory db=sqlite
2026-10-05T10:00:01 INFO  metric agent started    port=8083
2026-10-05T10:00:01 INFO  task manager started    port=8081
2026-10-05T10:00:01 INFO  scheduler started       port=8082 interval=10s
2026-10-05T10:00:01 INFO  worker started          port=8084 max_concurrent=3
2026-10-05T10:00:01 INFO  adaptive-cron is ready  gateway=8080
""",
        ),
        (
            "curl -s localhost:8083/api/v1/metrics/load",
            """
{
  "cpu_percent": 42.3,
  "memory_percent": 55.1,
  "load_avg_1": 1.24,
  "cores": 8,
  "load_score": 0.41
}
""",
        ),
        (
            "curl -s -X POST localhost:8082/api/v1/scheduler/trigger/42",
            """
{
  "task_id": "42",
  "decision": "deferred",
  "reason": "load_score 0.41 > max_load_score 0.30",
  "retry_in_seconds": 25
}
""",
        ),
    ],
)

# 12. API Endpoints
slide_table(
    "API Endpoints",
    ["Сервис", "Порт", "Основные endpoints"],
    [
        ["API Gateway", "8080", "GET /health, /api/v1/gateway/*"],
        ["Task Manager", "8081", "CRUD /api/v1/tasks"],
        ["Scheduler", "8082", "/api/v1/scheduler/next, /queue, /trigger/{id}"],
        ["Metric Agent", "8083", "/api/v1/metrics/current, /history, /load"],
        ["Worker", "8084", "/api/v1/execute, /status/{id}, /worker/stats"],
    ],
    accent=ACCENT,
    col_widths=[2, 1, 5],
)

# 13. Roadmap
slide_roadmap()

# 14. Финал
slide_final(
    "Adaptive Cron",
    [
        "Адаптивный планировщик, который не перегружает сервер",
        "",
        "Репозиторий: https://github.com/asvec6767/adaptive_cron",
        "Контакты: @datatape",
    ],
)

# ---------- Сохранение ----------
prs.save("adaptive_cron.pptx")
print("Готово: adaptive_cron.pptx — %d слайдов" % len(prs.slides))