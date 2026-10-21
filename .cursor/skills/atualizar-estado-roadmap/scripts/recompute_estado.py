#!/usr/bin/env python3
"""Recalcula o progresso do fluxo criminal completo a partir dos milestones."""

from __future__ import annotations

import re
from datetime import date
from pathlib import Path

CRIME_FLOW_THROUGH = 12  # M00..M12 = geração do caso até prisão do culpado
PLAYTEST_MILESTONE = 6  # M06 = primeiro recorte de playtest, não fim do jogo
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
        "O gerador determinístico v1 está em curso; o mundo ainda não está materializado."
    ),
    2: (
        "O gerador determinístico da cidade mínima está aceito. "
        "NPCs, empresas e relações persistidas estão em implementação."
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
        "com o servidor como autoridade. O fluxo investigativo completo ainda está em construção."
    ),
    7: (
        "O recorte web, a API e os registros bancários e do GATE estão aceitos. "
        "Correspondências procedurais ainda precisam entrar no ciclo investigativo."
    ),
    8: (
        "Registros financeiros, históricos do GATE e correspondências procedurais "
        "estão aceitos. Busca e apreensão ainda precisa conectar objetos do mundo aos casos."
    ),
    9: (
        "O jogador já pode consultar registros e correspondências e se deslocar "
        "entre cidades e pontos de serviço. Busca e apreensão ainda precisa conectar objetos aos casos."
    ),
    10: (
        "O jogador já pode guardar e organizar elementos por caso em whiteboards privados. "
        "A busca e apreensão e a acusação ainda precisam fechar o fluxo investigativo."
    ),
    11: (
        "Whiteboards e busca e apreensão estão aceitos. "
        "Acusação e julgamento ainda precisam fechar o fluxo criminal."
    ),
    12: (
        "O fluxo criminal completo está aceito: o caso é gerado, investigado, "
        "organizado e pode terminar na prisão do culpado."
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
    7: "Registros bancários e GATE",
    8: "Correio eletrônico",
    9: "Deslocamento e pontos de serviço",
    10: "Whiteboards pessoais",
    11: "Busca e apreensão",
    12: "Acusação e julgamento",
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
    path_items = [m for m in items if m["number"] <= CRIME_FLOW_THROUGH]
    path_items.sort(key=lambda m: m["number"])
    expected = list(range(0, CRIME_FLOW_THROUGH + 1))
    got = [m["number"] for m in path_items]
    if got != expected:
        raise SystemExit(
            f"Fluxo criminal esperado M00..M{CRIME_FLOW_THROUGH:02d}, "
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
    if milestone["number"] == PLAYTEST_MILESTONE:
        suffix = "   ← primeiro recorte de playtest"
    elif milestone["number"] == CRIME_FLOW_THROUGH:
        suffix = "   ← fluxo completo até a prisão"
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
    playtest_milestone = next(
        (m for m in milestones if m["number"] == PLAYTEST_MILESTONE), None
    )
    playtest_ready = bool(playtest_milestone and playtest_milestone["status"] == STATUS_ACEITO)
    crime_flow_done = bool(milestones and milestones[-1]["status"] == STATUS_ACEITO)
    vigente_cell = display_name(current) if current else "nenhum (fluxo criminal completo aceito)"
    playtest_cell = (
        "M06 aceito; recorte mínimo pronto para playtest interno."
        if playtest_ready
        else "Ainda não pronto; depende do aceite de M06."
    )
    if crime_flow_done:
        playtest_cell += " Fluxo completo até a prisão também aceito."

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

- **Alvo desta barra:** fluxo criminal completo, da geração do caso à prisão do culpado = **M{CRIME_FLOW_THROUGH:02d} aceito**.
- **Marco de playtest:** M{PLAYTEST_MILESTONE:02d} libera o primeiro recorte atravessável no browser; ele não representa o fluxo investigativo completo.
- **Fora desta barra:** Vesper, auth completa, simulação temporal geral e mapa rico — entram em milestones posteriores.

## Agora

| Campo | Valor |
| --- | --- |
| Data de referência | {today.isoformat()} |
| Milestone vigente | {vigente_cell} |
| Último aceito | {last_cell} |
| Primeiro playtest | {playtest_cell} |
| Fluxo do crime até a prisão | **{percent}%** ({accepted_n} de {total} milestones) |

{blurb}

## Barra do fluxo criminal completo

O fluxo começa com a geração determinística do mundo e do caso, passa pelas fontes de investigação e organização dos elementos, e termina quando uma acusação válida permite prender o culpado. O M{PLAYTEST_MILESTONE:02d} continua sendo um marco intermediário de playtest.

```text
{bar_block}
{pointer}
```

Cada milestone do fluxo pesa **1/{total} ≈ {weight}**. Só entra na conta com status **Aceito**. “Em curso” não soma porcentagem; a porcentagem mede ciclos aceitos do percurso completo.

{table}

Fórmula: `milestones_aceitos_M00_a_M{CRIME_FLOW_THROUGH:02d} / {CRIME_FLOW_THROUGH + 1}`, arredondado para inteiro. O primeiro playtest continua identificado pelo aceite do M06, em campo separado.

## Como atualizar

1. Marcar o milestone no próprio arquivo (`**Status:**` + checkboxes de aceite).
2. Rodar a skill / `python3 .cursor/skills/atualizar-estado-roadmap/scripts/recompute_estado.py`.
3. Se novos milestones entrarem no fluxo, atualizar `CRIME_FLOW_THROUGH`, a sequência em [`ciclos.md`](ciclos.md) e os rótulos/blurb deste script.
"""


def main() -> None:
    root = repo_root()
    milestones_dir = root / "docs" / "roadmap" / "milestones"
    estado_path = root / "docs" / "roadmap" / "estado.md"
    milestones = load_path_milestones(milestones_dir)
    estado_path.write_text(render(milestones, date.today()), encoding="utf-8")
    accepted = sum(1 for m in milestones if m["status"] == STATUS_ACEITO)
    playtest_status = "pronto" if any(
        item["number"] == PLAYTEST_MILESTONE and item["status"] == STATUS_ACEITO
        for item in milestones
    ) else "pendente"
    print(
        f"Atualizou {estado_path.relative_to(root)}: fluxo criminal "
        f"{accepted}/{len(milestones)} aceitos "
        f"({pct(accepted, len(milestones))}%); playtest M{PLAYTEST_MILESTONE:02d} {playtest_status}."
    )


if __name__ == "__main__":
    main()
