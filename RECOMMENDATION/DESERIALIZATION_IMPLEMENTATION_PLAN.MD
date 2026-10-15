# Deserializacja — fazy Z0–Z4

Status: plan. Wymagania: [DESERIALIZATION.MD](DESERIALIZATION.MD). Kontrakty: [ROADMAP.MD](ROADMAP.MD).

## Z0. Kontrakt modeli i ról

Zależności: B0–B3, CVE F1/F4, JVM G2–G3, DATA_FLOW D0–D2.

- Wprowadzić proponowany katalog `DeserializerModel`: komponent, zakres wersji, sygnatury wejścia, reguły typów, callbacki i obsługiwane zabezpieczenia.
- Zachować rolę entry point w istniejących ręcznych mapowaniach; nie podnosić starych mapowań XStream do potwierdzonych pełnych gadget chains.
- Zdefiniować osobne dowody: użycie API, kandydat typu, wywołanie callbacka, rozstrzygnięcie filtra i kandydat łańcucha.
- Uzgodnić z RTA rejestrowanie obiektów tworzonych przez deserializer bez jawnego `new`.

Gotowe: model rozróżnia podatny komponent, jego wejście i właściwy kod związany z danym CVE.

## Z1. Java serialization i callbacki

- Dodać model wybranych operacji ObjectInputStream i rozpoznawanie typów Serializable/Externalizable dostępnych w scenariuszu.
- Obsłużyć zadeklarowany zakres callbacków z dokładnymi sygnaturami i dostępem; prywatny `readObject` nie wymaga zwykłego publicznego wywołania.
- Odwzorować konstrukcję według reguł serializacji, nie przez wywołanie każdego konstruktora każdej klasy. Szczególne przypadki, np. rekordy i Externalizable, mają własny zakres pokrycia.
- Zbiór dopuszczalnych typów wynika z modelu i danych, a nie z bezwarunkowego połączenia wejścia ze wszystkimi metodami wszystkich klas Serializable.
- Brak znanego typu wejściowego może wymagać szerokiego zbioru kandydatów i oznaczonego ograniczenia; limit nie jest dowodem, że reszta typów nie występuje.

Gotowe: callback ma uzasadnioną ścieżkę, brak zgodnej sygnatury nie tworzy callbacka, a nieznany zbiór typów jest widoczny. To pierwszy przyrost mechanizmu.

## Z2. Filtry, konfiguracja i pierwszy model biblioteki

Zależności: DATA_FLOW D3, FRAMEWORK F1 w używanych scenariuszach konfiguracji.

- Powiązać filtr z konkretnym deserializatorem/kontekstem, określając priorytet źródeł w obsługiwanej wersji.
- Zaimplementować ograniczony katalog znanych filtrów/konfiguracji; dynamiczna polityka ma stan unknown.
- Wybrać jeden zakres biblioteki zewnętrznej, np. XStream w konkretnych wydaniach. Opisać źródła, reguły tworzenia typów i warunki zabezpieczeń właściwe temu produktowi.
- Nie przenosić reguł ObjectInputFilter automatycznie na biblioteki, które z nich nie korzystają w danym modelu.
- Zachować rozstrzygnięcia per kandydat/instancja; odrzucenie jednego typu nie oznacza wykluczenia całego CVE ani wszystkich innych łańcuchów.

Gotowe: filtr accept/reject/unknown, dwie konfiguracje deserializatora i różne wersje modelu są rozróżnione.

## Z3. Dane zewnętrzne i kandydaci łańcuchów

Zależności: DATA_FLOW D4, CVE F5 dla dowodów konkretnych CVE.

- Powiązać źródła danych z argumentem wejścia deserializacji; osobno pokazać brak analizy flow i znalezioną ścieżkę.
- Wprowadzić mały wersjonowany katalog kandydatów łańcuchów z wymaganymi klasami, wersjami, warunkami pól i konfiguracji. Dla nierozstrzygniętych przesłanek zachować warunki.
- Sprawdzać zgodność obecnego bytecode i tożsamości komponentów. Związek po nazwie klasy nie wystarcza dla shaded lub przesłoniętej kopii.
- Rozdzielić wyniki `candidate_chain`, `conditions_unresolved`, `candidate_blocked` od potwierdzenia wykorzystania. Nie oznaczać całej podatności jako nieobecnej po zablokowaniu jednego znanego łańcucha.
- Określić budżet wyszukiwania i zachować informację o niekompletności; nie twierdzić, że katalog zawiera wszystkie łańcuchy.

Gotowe: wymagana brakująca klasa, zablokowany kandydat, nieznany filtr i znaleziony dopływ danych mają osobne uzasadnienia. Nie jest to pełny dowód exploitability.

## Z4. Kwalifikacja i raportowanie

- Przygotować nieszkodliwe fixtures demonstrujące callbacki i przepływ sterowania, z zapisanym oczekiwanym modelem, oraz osobny źródłowo opisany zestaw CVE.
- Pokazać w raporcie wejście, wersję modelu, filtry i przesłanki łańcucha; snapshot obejmuje wszystkie użyte reguły i konfiguracje.
- Sprawdzić scenariusze: private callback, Externalizable, szczególny typ bez pokrycia, dwie instancje, filtr i nieznany zbiór typów.
- Wykonać kontrole B4, porównać wyniki z ręcznie udokumentowanymi przypadkami i nie uznawać wzrostu liczby CVE za dowód poprawy.
