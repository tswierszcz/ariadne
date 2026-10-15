# Inne ekosystemy i pilotaż Go — rekomendacja funkcjonalna

Status: plan eksploracyjny. Fazy: [ECOSYSTEMS_IMPLEMENTATION_PLAN.MD](ECOSYSTEMS_IMPLEMENTATION_PLAN.MD). Model wspólny: [ROADMAP.MD](ROADMAP.MD).

## Cel i stan obecny

Umożliwić wspólne raportowanie podatności z innego ekosystemu, zaczynając od małego pilota Go opartego na istniejącym narzędziu. Nie przenosić założeń klas JVM na język o innym modelu wykonania.

Obecny pipeline jest zależny od GAV, bytecode klas, hierarchii i profili JVM. Nawet `internal/reach` korzysta z hierarchii i GAV, a nie wyłącznie z abstrakcyjnego BFS. Współdzielenie wymaga wyodrębnienia kontraktów; nie jest prostym włączeniem flagi ecosystem.

## Pierwsza funkcjonalność: import wyniku Go

- Przyjąć ustrukturyzowany wynik `govulncheck` dla przypiętej wersji formatu/narzędzia i pokazać go we wspólnym raporcie.
- Zachować identyfikatory advisory, moduły, pakiety/symbole, informacje o wywołaniach oraz poziom analizy źródłowej lub binarnej.
- Zachować oryginalne znaczenie i dowody. Znaleziony symbol w binarium nie jest automatycznie dowodem ścieżki od wejścia aplikacji.
- Nie nadawać Go statusu `FRAMEWORK` przez zastosowanie list pakietów Spring ani kopiować ograniczeń JDK.
- Podać brakujące dane i ograniczenia dostawcy. Nieobsługiwany schemat kończy import czytelnym błędem.

## Opcjonalne uruchomienie narzędzia

Po imporcie można dodać jawne uruchomienie lokalnej, przypiętej wersji `govulncheck` jako osobną komendę/adapter. Wejście obejmuje źródła lub binarium, build context, wersję Go i konfigurację bazy danych. Nie instalować ani nie pobierać innej wersji narzędzia w tle jako ubocznego efektu raportowania.

Dla trybu źródłowego trzeba uwzględnić m.in. workspace, moduły, replace, build tags i docelową platformę. Lista `go.mod`/`go.sum` nie jest automatycznie dokładnym wykazem wszystkich komponentów faktycznie użytych w danym binarium.

## Wyniki, zakres i granice

Tryby źródłowy i binarny mają różną precyzję. Wspólny wynik przechowuje oryginalny poziom dowodu i zakres analizy, zamiast wymuszać równoważność z JVM REACHABLE. Negatywny wynik zachowuje ograniczenia narzędzia, konfiguracji i danych podatności.

Inne języki wymagają osobnych frontendów i adapterów. Wspólny model komponentów, dowodów, raportu i snapshotów można współdzielić; parser klas, reguły inicjalizacji, graf i profile nie przechodzą bezpośrednio między ekosystemami.

MVP to importer znanego formatu Go. Samodzielny analizator Go oraz automatyczne wsparcie dowolnego języka pozostają poza pierwszym zakresem.

## Gotowość i źródła

Gotowość: źródłowy i binarny wynik Go zachowują różną siłę dowodów, import działa offline na zachowanym raporcie, a nieznane pola/wersja są obsługiwane zgodnie z jawnym kontraktem.

- [Go: zarządzanie podatnościami](https://go.dev/doc/security/vuln/).
- [govulncheck: tryby i ograniczenia](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck).

Implementacja ma przypiąć konkretną wersję narzędzia; bieżący opis nie jest obietnicą niezmienności zewnętrznego formatu.
