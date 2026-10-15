# Start i kontynuacja implementacji w nowej sesji

Skill: [ariadne-implement](skills/ariadne-implement/SKILL.md). Rejestr postępu: [IMPLEMENTATION_STATUS.MD](IMPLEMENTATION_STATUS.MD). Aktualny zakres produktu: [SUPPORT.MD](../docs/SUPPORT.MD).

W tym środowisku skill został zarejestrowany przez odnośnik
`~/.agents/skills/ariadne-implement` do wersjonowanej kopii w repozytorium.
Zmiany instrukcji w tej kopii nie wymagają ręcznego synchronizowania drugiej kopii.
Po przeniesieniu repozytorium na inne stanowisko nadal można użyć promptu z jawną
ścieżką opisanego niżej; lokalna rejestracja nie jest częścią checkoutu Git.

## Prompt startowy

Wklej w nowej sesji otwartej na repozytorium Ariadne:

```text
Użyj $ariadne-implement do realizacji planów Ariadne z RECOMMENDATION.

Przeczytaj indeks, ROADMAP, IMPLEMENTATION_STATUS i docs/SUPPORT.MD.
Sprawdź stan kodu, wznowienie i zależności w trybie tylko do odczytu.
Następnie ponownie przedstaw plan: fazy, kolejność, priorytety,
konkretny zakres przyrostu, oczekiwany rezultat, weryfikację i dokumentację.
Jeśli jest to pierwsza sesja, zaproponuj fundamenty ROADMAP:R0
i minimalne wspólne kontrakty B0–B1.

Zaczekaj na moją wyraźną zgodę, żebym mógł zmienić kolejność lub priorytety.
Przed zgodą nie zmieniaj plików, nie implementuj i nie uruchamiaj testów.
Po zatwierdzeniu wykonaj zaakceptowany zakres, jego weryfikację
i zapis stanu do kontynuacji.

Warunkiem ukończenia jest aktualizacja opisu faktycznie wspieranych funkcji:
docs/SUPPORT.MD, odpowiednich sekcji README, a gdy zmienia się ich treść —
także ograniczeń generowanego raportu i właściwych FAQ.
Opisz precyzyjnie obsługiwane mechanizmy, API i wersje, warunki działania,
sposób włączenia, dowody weryfikacji i pozostałe ograniczenia.
Nie przedstawiaj zaplanowanej lub częściowej funkcji jako pełnego wsparcia.

Na końcu zaktualizuj IMPLEMENTATION_STATUS, wskaż ukończoną fazę,
co jest teraz wspierane oraz dokładny następny krok dla kolejnej sesji.
Jeśli konieczna decyzja nie wynika z kodu ani planu, zapytaj konkretnie.
Kontynuuj niezależną część pracy tylko w ramach już zatwierdzonego zakresu.
```

Każda nowa sesja zaczyna się ponownym przedstawieniem planu i oczekiwaniem na zgodę. Domyślny zakres to jeden przyrost na wywołanie; można zatwierdzić kilka faz wraz z kolejnością. Faza kontraktowa daje trwały kontrakt i scenariusze, nie deklarację gotowej funkcji użytkowej.

## Wybrany obszar i faza

```text
Użyj $ariadne-implement. Zrealizuj JVM_GRAPH:G2 zgodnie z planem tego obszaru.
Najpierw sprawdź zależności i dotychczasowe dowody tylko do odczytu.
Przedstaw plan, priorytety i konieczne zależności, a następnie zaczekaj
na moją zgodę. Dopiero po zatwierdzeniu wykonaj implementację,
weryfikację oraz aktualizację opisu wsparcia i rejestru postępu.
```

Używaj pełnego identyfikatora: np. `CVE:F4`, `FRAMEWORKS:F2`, `REFLECTION:R1`. Same oznaczenia `F2`/`R0` występują w kilku dokumentach.

## Kontynuacja

```text
Użyj $ariadne-implement i kontynuuj od stanu w
RECOMMENDATION/IMPLEMENTATION_STATUS.MD. Najpierw porównaj zapis z kodem
i wykonanymi kontrolami w trybie tylko do odczytu. Ponownie przedstaw plan
wznowienia lub kolejnego przyrostu, kolejność i priorytety. Zaczekaj na moją
zgodę przed realizacją. Po zatwierdzeniu wykonaj plan wraz z aktualizacją
dokumentacji wspieranych możliwości.
```

## Gdy skill nie jest dostępny w selektorze

Wersjonowana kopia znajduje się w repozytorium. Ten prompt działa także bez automatycznego wykrycia skilla:

```text
Przeczytaj RECOMMENDATION/skills/ariadne-implement/SKILL.md
i zastosuj opisany tam workflow do tego repozytorium.
Najpierw przedstaw plan następnej gotowej fazy z RECOMMENDATION,
kolejność i priorytety; zaczekaj na moją wyraźną zgodę.
Po zatwierdzeniu zrealizuj uzgodniony zakres wraz z weryfikacją,
dokumentacją faktycznego wsparcia i zapisem postępu.
```

Lokalne wykrywanie skilli i jawne wywołanie opisuje [oficjalna dokumentacja OpenAI](https://learn.chatgpt.com/docs/build-skills). Wersjonowana kopia pozostaje źródłem instrukcji projektu; instalacja powinna wskazywać tę kopię albo być aktualizowana wraz z nią.
