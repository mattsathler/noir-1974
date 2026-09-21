#!/usr/bin/env python3
"""Recalcula docs/roadmap/estado.md a partir do Status dos milestones."""

from __future__ import annotations

import re
from datetime import date
from pathlib import Path

PLAYTEST_THROUGH = 6  # M00..M06 = caminho até o 1º deploy/playtest
BAR_WIDTH = 20

STATUS_ACEITO = "aceito"
STATUS_EM_CURSO = "em curso"
STATUS_NAO_INICIADO = "não iniciado"

PRODUCT_BLURB = {
    -1: (
        "Ainda não há milestone aceito. O caminho até o playtest não começou."
    ),
    0: (
        "O produto, hoje, é o repositório e a documentação. "
        "`apps/web` e `apps/api` existem só como fronteira do monorepo."
    ),
    1: (
        "Modelo de entidades e schema inicial estão escritos. "
        "Ainda não há gerador nem mundo materializado."
    ),
    2: (
        "Há gerador determinístico da cidade mínima. "
        "Sociedade (NPCs, empresas, relações) ainda não está no recorte aceito."
    ),
    3: (
        "Há sociedade mínima: NPCs, empresas e relações consultáveis. "
        "Mutations persistentes ainda não estão no recorte aceito."
    ),
    4: (
        "O mundo reconstruído sobrevive a restart com mutations. "
        "Ainda não há gerador de casos aceito."
    ),
    5: (
        "Há gerador de casos simples com validação. "
        "Ainda não há API pública nem cliente web do playtest."
    ),
    6: (
        "Primeiro recorte atravessável no browser: mundo, NPC/relações e caso, "
        "com o servidor como autoridade. Pronto para deploy e playtest interno."
    ),
}

TABLE_LABELS = {
    0: "Repositório e docs",
    1: "Entidades e schema",
    2: "Gerador determinístico",
    3: "Sociedade e relações",
    4: "Mutations e reconstrução",
    5: "Casos e validação",
    6: "API e cliente web",
}

HEADING_RE = re.compile(
    r"^#\s+(M(\d+))\s+[—–-]\s+(.+)$", re.MULTILINE
)
STATUS_RE = re.compile(
    r"^\*\*Status:\*\*\s*(.+?)\s*$", re.MULTILINE
)


def repo_root() -> Path:
    here = Path(__file__).resolve()
    for parent in here.parents:
        if (parent / "docs" / "roadmap" / "milestones").is_dir():
            return parent
    raise SystemExit("Não achei docs/roadmap/milestones a partir do script.")


def normalize_status(raw: str) -> str:
    value = raw.strip().rstrip(".").lower()
    aliases = {
        "aceito": STATUS_ACEITO,
        "aceita": STATUS_ACEITO,
        "done": STATUS_ACEITO,
        "em curso": STATUS_EM_CURSO,
        "em andamento": STATUS_EM_CURSO,
        "não iniciado": STATUS_NAO_INICIADO,
        "nao iniciado": STATUS_NAO_INICIADO,
    }
    if value not in aliases:
        raise SystemExit(f"Status desconhecido: {raw!r}")
    return aliases[value]


def parse_milestone(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")
    heading = HEADING_RE.search(text)
    status = STATUS_RE.search(text)
    if not heading or not status:
        raise SystemExit(f"Milestone inválido (H1 Mnn + Status): {path}")
    number = int(heading.group(2))
    return {
        "id": heading.group(1),
        "number": number,
        "title": heading.group(3).strip(),
        "status": normalize_status(status.group(1)),
        "filename": path.name,
    }


def load_path_milestones(milestones_dir: Path) -> list[dict]:
    files = sorted(milestones_dir.glob("M*.md"))
    items = [parse_milestone(path) for path in files]
    path_items = [m for m in items if m["number"] <= PLAYTEST_THROUGH]
    path_items.sort(key=lambda m: m["number"])
    expected = list(range(0, PLAYTEST_THROUGH + 1))
    got = [m["number"] for m in path_items]
    if got != expected:
        raise SystemExit(
            f"Caminho do playtest esperado M00..M{PLAYTEST_THROUGH:02d}, "
            f"achei {got}"
        )
    return path_items


def pct(accepted: int, total: int) -> int:
    return round(accepted / total * 100)


def weight_label(total: int) -> str:
    return f"{(100 / total):.1f}%".replace(".", ",")


def bar_line(milestone: dict) -> str:
    filled = BAR_WIDTH if milestone["status"] == STATUS_ACEITO else 0
    bar = "█" * filled + "░" * (BAR_WIDTH - filled)
    suffix = ""
    if milestone["number"] == PLAYTEST_THROUGH:
        suffix = "   ← deploy / playtest"
    return f"{milestone['id']} {bar}  {milestone['status']}{suffix}"


def vigente(milestones: list[dict]) -> dict | None:
    for item in milestones:
        if item["status"] != STATUS_ACEITO:
            return item
    return None


def last_accepted(milestones: list[dict]) -> dict | None:
    accepted = [m for m in milestones if m["status"] == STATUS_ACEITO]
    return accepted[-1] if accepted else None


def display_name(milestone: dict) -> str:
    label = TABLE_LABELS.get(milestone["number"])
    if label:
        return f"{milestone['id']} — {label}"
    return f"{milestone['id']} — {milestone['title']}"


def render(milestones: list[dict], today: date) -> str:
    total = len(milestones)
    accepted_n = sum(1 for m in milestones if m["status"] == STATUS_ACEITO)
    percent = pct(accepted_n, total)
    last = last_accepted(milestones)
    current = vigente(milestones)
    last_n = last["number"] if last else -1
    playtest_done = (
        milestones[-1]["status"] == STATUS_ACEITO
        if milestones
        else False
    )

    if playtest_done:
        vigente_cell = "nenhum (caminho do playtest aceito)"
        deploy_cell = (
            "Sim. Recorte M06 atravessável no browser "
            "(deploy / playtest interno)."
        )
    elif current:
        vigente_cell = display_name(current)
        deploy_cell = (
            "Não. Não há servidor, cliente jogável nem mundo gerado."
            if last_n < 2
            else "Não. M06 ainda não foi aceito."
        )
    else:
        vigente_cell = "—"
        deploy_cell = "Não."

    last_cell = display_name(last) if last else "nenhum"
    blurb = PRODUCT_BLURB.get(last_n, PRODUCT_BLURB[-1])
    weight = weight_label(total)

    bar_block = "\n".join(bar_line(m) for m in milestones)
    pointer = " " * 5 + f"| {percent}%"

    rows = []
    for index, item in enumerate(milestones):
        cumulative = pct(index + 1, total)
        cumulative_cell = f"**{cumulative}%**" if index == total - 1 else f"{cumulative}%"
        status_cell = item["status"].capitalize() if item["status"] != STATUS_NAO_INICIADO else "Não iniciado"
        if item["status"] == STATUS_ACEITO:
            status_cell = "Aceito"
        elif item["status"] == STATUS_EM_CURSO:
            status_cell = "Em curso"
        label = TABLE_LABELS.get(item["number"], item["title"])
        link = f"[{item['id']}](milestones/{item['filename']}) {label}"
        rows.append(f"| {link} | {weight} | {status_cell} | {cumulative_cell} |")

    table = "\n".join(
        [
            "| Milestone | Peso | Status | Soma acumulada se aceito |",
            "| --- | ---: | --- | ---: |",
            *rows,
        ]
    )

    return f"""# Estado do produto

Fonte única do **onde estamos**. Regenerar com a skill `atualizar-estado-roadmap`
(script `scripts/recompute_estado.py`). Não editar a porcentagem à mão.

- **Alvo desta barra:** primeiro deploy + playtest interno = **M{PLAYTEST_THROUGH:02d} aceito**.
- **Fora desta barra:** Vesper, auth completa, simulação temporal, mapa rico — entram em milestones posteriores.

## Agora

| Campo | Valor |
| --- | --- |
| Data de referência | {today.isoformat()} |
| Milestone vigente | {vigente_cell} |
| Último aceito | {last_cell} |
| Deploy / playtest | {deploy_cell} |
| Caminho até o 1º playtest | **{percent}%** ({accepted_n} de {total} milestones) |

{blurb}

## Barra até o primeiro deploy e playtest

Playtest = recorte M{PLAYTEST_THROUGH:02d} no browser: consultar mundo, um NPC com relações e um caso, com o servidor como autoridade.

```text
{bar_block}
{pointer}
```

Cada milestone do caminho pesa **1/{total} ≈ {weight}**. Só entra na conta com status **Aceito**. “Em curso” não soma porcentagem (evita 40% de trabalho pela metade).

{table}

Fórmula: `aceitos_no_caminho / {total}`, arredondado para inteiro.

## Como atualizar

1. Marcar o milestone no próprio arquivo (`**Status:**` + checkboxes de aceite).
2. Rodar a skill / `python3 .cursor/skills/atualizar-estado-roadmap/scripts/recompute_estado.py`.
3. Se o primeiro playtest ganhar milestones extras **antes** de M{PLAYTEST_THROUGH:02d}, ajustar `PLAYTEST_THROUGH` no script e [`ciclos.md`](ciclos.md).
"""


def main() -> None:
    root = repo_root()
    milestones_dir = root / "docs" / "roadmap" / "milestones"
    estado_path = root / "docs" / "roadmap" / "estado.md"
    milestones = load_path_milestones(milestones_dir)
    estado_path.write_text(render(milestones, date.today()), encoding="utf-8")
    accepted = sum(1 for m in milestones if m["status"] == STATUS_ACEITO)
    print(
        f"Atualizou {estado_path.relative_to(root)}: "
        f"{accepted}/{len(milestones)} aceitos "
        f"({pct(accepted, len(milestones))}%)."
    )


if __name__ == "__main__":
    main()
