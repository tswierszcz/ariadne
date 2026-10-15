# Stan realizacji i przekazanie między sesjami

Ten plik jest rejestrem pracy nad nowymi planami, a nie deklaracją wsparcia produktu. Aktualne możliwości opisuje [SUPPORT.MD](../docs/SUPPORT.MD), zakres prac [ROADMAP.MD](ROADMAP.MD), a sposób prowadzenia sesji [skill](skills/ariadne-implement/SKILL.md).

## Stan początkowy

Przygotowano plany, prompt i skill. W ramach ich tworzenia nie rozpoczęto implementacji nowych faz. Istniejący kod może spełniać część kryteriów; każda kolejna sesja ma to zweryfikować zamiast zakładać, że wszystko trzeba napisać od początku.

Stany faz: `UNASSESSED`, `READY`, `IN_PROGRESS`, `NEEDS_VERIFICATION`, `DONE`, `BLOCKED`. `DONE` oznacza spełnienie wszystkich kryteriów konkretnej fazy oraz wymaganej weryfikacji i dokumentacji. `BLOCKED` wymaga konkretnego powodu i sposobu odblokowania. To etykiety tego pliku, nie stany celów w aplikacji Codex.

## Przegląd obszarów

| Obszar | Plan | Ocena nowych faz | Wznowienie |
|---|---|---|---|
| ROADMAP | [Kontrakty B0–B4 i przyrosty](ROADMAP.MD) | UNASSESSED | Uzgodnić B0–B1 z CVE/JVM; nie traktować całego R0 jako ukończonego po samym projekcie. |
| CVE | [F0–F9](CVE_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Sprawdzić wspólne wymagania F0–F1 i aktualny model wyników. |
| JVM_GRAPH | [G0–G5](JVM_GRAPH_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Po minimalnym B0–B1 dobrać pierwszy wykonalny przyrost; sprawdzić istniejący graf. |
| DATA_FLOW | [D0–D5](DATA_FLOW_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Wymaga odpowiednich fundamentów JVM; nie zakładać istnienia CFG/value analysis. |
| SERVICE_LOADING | [S0–S4](SERVICE_LOADING_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Rejestry usług i lokalne wartości według zależności planu. |
| REFLECTION | [R0–R4](REFLECTION_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Wspólne wartości DATA_FLOW i reguły rozwiązywania JVM. |
| FRAMEWORKS | [F0–F5](FRAMEWORKS_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Wykorzystać istniejące profile; nie mylić ich z pełnym modelem runtime’u. |
| ASYNC_CALLBACKS | [A0–A4](ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Katalog API i powiązanie callbacka z właściwym obiektem. |
| DYNAMIC_LINKAGE | [L0–L4](DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Uporządkować istniejące lambda/method references przed szerszymi uchwytami. |
| DESERIALIZATION | [Z0–Z4](DESERIALIZATION_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Role entry point/mapowanie CVE i wspólny model obiektów. |
| GRAPH_PRECISION | [P0–P4](GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Najpierw poprawność dispatch i alokacje, potem RTA. |
| RUNTIME_NATIVE | [N0–N5](RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Importer i kontrakt dowodów przed opcjonalnym agentem. |
| ECOSYSTEMS | [E0–E4](ECOSYSTEMS_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Osobny pilot; nie jest warunkiem ulepszania JVM. |

Po pierwszej sesji wpisz poniżej konkretne identyfikatory faz i ich stan. Ogólne `UNASSESSED` w tabeli obszarów zastąp krótkim podsumowaniem rzeczywistego postępu; nie oznaczaj całego obszaru DONE po jednej fazie.

## Rejestr faz

Na razie nie zakwalifikowano żadnej nowej fazy jako DONE. Przy dodawaniu wpisu zapisz:

- Pełny identyfikator fazy i stan.
- Zakres zlecenia i niezbędne wspólne zależności, które faktycznie wykonano.
- Zmienione pliki i zachowanie dostępne użytkownikowi albo rezultat kontraktowy.
- Spełnione i niespełnione kryteria, z odnośnikami do dowodów.
- Polecenia rzeczywiście wykonane, ich wyniki i zakres; osobno kontrole niewykonane z powodem.
- Zmiany w SUPPORT, README, raporcie i innych dotkniętych opisach, albo uzasadnienie braku zmiany zakresu wsparcia.
- Pozostałe ograniczenia, konieczne decyzje oraz następny krok.

## Punkt startowy następnej sesji

1. Porównaj ten stan z bieżącym kodem i zmianami w checkoutcie.
2. Przeczytaj ROADMAP:R0, B0–B1 oraz właściwe części CVE:F0/F1 i JVM_GRAPH:G0.
3. Zaproponuj najbliższą wykonalną fazę lub minimalny wspólny kontrakt, przedstaw zakres, kolejność i priorytety, a następnie zaczekaj na wyraźną zgodę użytkownika. Do tego czasu wykonuj tylko przegląd do odczytu.
4. Po zatwierdzeniu wykonaj rezultat fazy, jego weryfikację i aktualizację dokumentacji. Jeżeli jest to faza kontraktowa, nie oznaczaj przez to funkcji jako zaimplementowanej.

Wersjonowane dowody i stan repozytorium są podstawą wznowienia. Poprzedni opis czatu nie zastępuje ich weryfikacji.
