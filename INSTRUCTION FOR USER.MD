# Instrukcja dla użytkownika — praca nad rozwojem Ariadne

> **Dokument wyłącznie dla użytkownika.** AI ma pomijać ten plik przy starcie sesji: nie odczytywać go automatycznie, nie analizować, nie streszczać ani nie dołączać do kontekstu inicjalizacyjnego. Odczyt jest dopuszczalny tylko na wyraźną prośbę użytkownika dotyczącą tego poradnika. Instrukcje operacyjne dla AI znajdują się w skillu `ariadne-implement`.

## 1. Zwykły przebieg pracy

Najwygodniejszy cykl to:

```text
Nowa sesja → propozycja planu → Twoja zgoda → realizacja fazy
          → weryfikacja → aktualizacja dokumentacji i zapis postępu
```

Nie musisz za każdym razem przypominać całej historii projektu. Agent powinien odczytać zapisane plany i postęp oraz porównać je z aktualnym kodem.

## 2. Rozpocznij nową sesję

Otwórz nowy chat w projekcie Ariadne, korzystający z tego samego katalogu repozytorium. Wklej:

```text
Użyj $ariadne-implement. Sprawdź aktualny postęp i zaproponuj
następny przyrost zgodnie z ROADMAP. Przedstaw zakres, kolejność,
priorytety i sposób weryfikacji. Zaczekaj na moją zgodę.
```

Skill jest przygotowany do prowadzenia tego workflow. Przed zgodą agent ma jedynie przeglądać kod i dokumentację. Samo jego wywołanie nie jest zatwierdzeniem implementacji.

## 3. Oceń plan i zmień priorytety, jeśli potrzebujesz

Agent powinien przedstawić konkretny zakres, zależności i oczekiwany rezultat. Możesz odpowiedzieć:

```text
Najpierw chciałbym poprawić mapowanie CVE.
Zaproponuj odpowiedni zakres i wyjaśnij jego zależności.
```

Możesz również ograniczyć zakres sesji:

```text
W tej sesji chcę tylko jedną fazę. Nie rozszerzaj zakresu
o inne mechanizmy, chyba że są konieczną zależnością.
```

Agent powinien przedstawić zmieniony plan i zaczekać na jego zatwierdzenie. Nie musi ignorować zależności technicznych: powinien wyjaśnić, które prace są konieczne przed realizacją Twojego priorytetu.

## 4. Zatwierdź konkretny plan

Gdy zakres Ci odpowiada, napisz:

```text
Zatwierdzam przedstawiony plan. Start.
```

Od tego momentu agent wykonuje zaakceptowaną implementację, weryfikację i dokumentację. Nie powinien pytać o każdy rutynowy krok. Istotna zmiana zakresu, kolejności lub priorytetów wymaga ponownego uzgodnienia.

Możesz zatwierdzić kilka faz i ich kolejność. Na początek wygodne są pojedyncze, zamknięte przyrosty, których efekt łatwo ocenić.

## 5. Odbierz rezultat jako zmianę możliwości systemu

Na zakończenie oczekuj odpowiedzi na pięć pytań:

- Co zostało ukończone?
- Co Ariadne teraz wspiera i pod jakimi warunkami?
- Jak zostało to sprawdzone?
- Jakie ograniczenia pozostały?
- Co jest następnym krokiem?

Sprawdzaj, czy opis wsparcia jest konkretny. „Obsługiwany ServiceLoader na classpath dla wskazanych operacji” jest bardziej użyteczne niż samo „dodano ServiceLoader”.

Jeżeli wymagane kontrole nie zostały wykonane, faza powinna mieć stan `NEEDS_VERIFICATION`, zamiast być przedstawiana jako ukończona. Faza projektu kontraktu może być ukończona, mimo że funkcja użytkowa nadal nie jest zaimplementowana — agent powinien jasno rozróżniać te rezultaty.

## 6. Kontynuuj w kolejnej sesji

Wklej:

```text
Użyj $ariadne-implement. Kontynuuj na podstawie
RECOMMENDATION/IMPLEMENTATION_STATUS.MD.
Sprawdź zapis względem kodu, przedstaw plan dalszych prac
i zaczekaj na moją zgodę.
```

Agent ma odczytać zapisany postęp. Nie musisz kopiować podsumowań z poprzedniego czatu. Każda nowa sesja ponownie przedstawia plan i czeka na zgodę, abyś mógł zmienić priorytety.

## 7. Do czego służą dokumenty

| Gdy chcesz… | Otwórz |
|---|---|
| Zobaczyć wszystkie obszary rozwoju | [Indeks rekomendacji](RECOMMENDATION/README.MD) |
| Zrozumieć kolejność i zależności | [ROADMAP](RECOMMENDATION/ROADMAP.MD) |
| Sprawdzić, gdzie zakończyła się praca | [IMPLEMENTATION_STATUS](RECOMMENDATION/IMPLEMENTATION_STATUS.MD) |
| Dowiedzieć się, co działa obecnie i jakie ma ograniczenia | [SUPPORT](docs/SUPPORT.MD) |
| Znaleźć pełne prompty i warianty uruchomienia | [START_IMPLEMENTATION](RECOMMENDATION/START_IMPLEMENTATION.MD) |
| Wrócić do ustaleń audytu | [AUDIT](AUDIT.MD) |
| Sprawdzić instrukcje prowadzące agenta | [Skill ariadne-implement](RECOMMENDATION/skills/ariadne-implement/SKILL.md) |

**Rekomendacje opisują zamierzenia, rejestr postępu opisuje wykonane prace, a SUPPORT opisuje aktualnie dostępne możliwości produktu.** Zapisanie planu nie oznacza wdrożenia funkcji.

## 8. Wybierz konkretną fazę albo opisz potrzebę

Jeśli wskazujesz fazę, używaj pełnego identyfikatora, np. `JVM_GRAPH:G2`, `CVE:F4` lub `FRAMEWORKS:F2`. Same oznaczenia `F2` i `R0` występują w kilku dokumentach.

Możesz też napisać zwyczajnie:

```text
Użyj $ariadne-implement. Teraz zależy mi na poprawności mapowań CVE.
Dobierz odpowiednie fazy, wyjaśnij zależności i przedstaw plan
do mojego zatwierdzenia.
```

Dobór faz i wyjaśnienie zależności należą wtedy do agenta. Nie musisz sam ustalać kolejności technicznych refaktorów.

## 9. Gdy skill nie jest widoczny

Użyj jawnej ścieżki do instrukcji:

```text
Przeczytaj RECOMMENDATION/skills/ariadne-implement/SKILL.md
i zastosuj jego instrukcje. Przedstaw plan i zaczekaj na moją zgodę.
```

W tym środowisku lokalna rejestracja skilla wskazuje wersjonowaną kopię w repozytorium. Przy przeniesieniu projektu na inne stanowisko lokalny odnośnik nie przenosi się automatycznie; prompt z jawną ścieżką nadal pozwala skorzystać z instrukcji.

## 10. Zachowuj dokumenty razem z projektem

Nowy chat w tym samym checkoutcie zobaczy zapisane pliki. W innym worktree lub na innym komputerze muszą być dostępne dokumenty, skill i aktualny rejestr postępu. Samo istnienie poprzedniego czatu tego nie zapewnia.

Warto zachowywać zmiany kodu i odpowiadającej im dokumentacji w kontroli wersji. Agent powinien przed rozpoczęciem rozpoznać istniejące zmiany użytkownika i uwzględnić je w proponowanym planie.

## Źródło informacji o wywołaniu skilli

[Oficjalna dokumentacja OpenAI: używanie i wykrywanie skilli](https://learn.chatgpt.com/docs/build-skills).
