# M02 — Gerador determinístico (cidade mínima)

**Ciclo:** 2  
**Status:** Aceito
**Depende de:** M01 aceito

## Objetivo

Um gerador versionado produz, a partir da world seed, um recorte de cidade (geografia/locais e, se o schema já exigir, esboço de empresas) de forma repetível.

## Critérios de aceite

- [x] Existe `generator_version` (começar em `1`).
- [x] Mesma seed + versão 1 gera o mesmo conjunto de entidades do recorte, em teste automatizado, duas vezes seguidas e após reiniciar o processo.
- [x] RNG apenas derivado da seed.
- [x] IDs estáveis entre as duas execuções.
- [x] Documentado o que a versão 1 gera e o que ainda não gera.

## Fora de escopo

Famílias completas, casos, HTTP, mutations.
