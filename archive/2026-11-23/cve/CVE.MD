# Recommendation CVE — plan funkcjonalny rozpoznawania podatności

Status: propozycja do realizacji. Dokument nie opisuje już wdrożonych funkcji.

Plan implementacyjny: [CVE_IMPLEMENTATION_PLAN.MD](CVE_IMPLEMENTATION_PLAN.MD).

## 1. Cel i wspólna podstawa

Rozszerzenie obejmuje pięć obszarów:

1. Mapowania CVE zależne od wersji biblioteki i dokładnej sygnatury metody.
2. Analizę kodu przed poprawką, w tym metod usuniętych i przemianowanych.
3. Rozpoznawanie bibliotek osadzonych w shaded JAR-ach.
4. Jawną świeżość danych CVE, odświeżanie cache i odtwarzanie analizy.
5. Podatności faktycznie używanego JDK/JRE.

Każdy wynik powinien odpowiadać na trzy pytania: czy rozpoznano komponent i jego wersję, czy istnieje wiarygodne mapowanie CVE na kod tej wersji oraz czy ten kod jest osiągalny w analizowanej aplikacji.

Obecnie resolver przechowuje listę metod dla CVE. Docelową jednostką wyniku będzie:

```text
CVE + komponent + wersja + konkretna kopia komponentu
```

Pozwoli to odróżnić np. zwykły JAR biblioteki od jej starszej kopii osadzonej w innym JAR-ze.

| Informacja | Przykładowa zawartość |
|---|---|
| Tożsamość komponentu | Współrzędne Maven, wersja, hash artefaktu. |
| Lokalizacja | Osobny JAR lub komponent osadzony w innym artefakcie. |
| Podatność wersji | Dotyczy / nie dotyczy / nierozstrzygnięta. |
| Mapowanie na kod | Dokładne / przybliżone / brak / konflikt. |
| Podstawa mapowania | Wpis ręczny, analiza poprawki, sugestia LLM. |
| Osiągalność | Status Ariadne albo brak możliwości oceny. |
| Dowody | Advisory, commit, plik, metoda, daty danych i uzasadnienie. |

Dokładne dopasowanie sygnatury nie oznacza potwierdzenia, że metoda jest podatna. Precyzja dopasowania i podstawa merytoryczna mapowania pozostają oddzielnymi informacjami. Podobnie znalezienie wpisu w metadanych nie dowodzi obecności odpowiadającego mu kodu.

## 2. Mapowania zależne od wersji i sygnatury

### Cel

Ariadne stosuje mapowanie właściwe dla analizowanej wersji i rozróżnia przeciążenia metod.

### Zachowanie

- Wpis mapowania wskazuje CVE, komponent, zakres wersji oraz metody.
- Metoda może mieć pełny deskryptor JVM, obejmujący parametry i typ wyniku.
- Różne zakresy wersji mogą mieć różne mapowania tego samego CVE.
- Dopasowanie sprawdza, czy metoda należy do właściwego komponentu i kopii artefaktu.
- Stare wpisy `Class#method` pozostają obsługiwane jako mapowania przybliżone.
- Rola metody jest jawna: kod podatny albo punkt wejścia do mechanizmu podatności. Nie są one traktowane jako równoważne dowody.

Przykład: biblioteka ma `parse(String)` oraz `parse(InputStream)`. Jeśli mapowanie wskazuje tylko pierwsze przeciążenie, wywołanie drugiego nie powinno samo wystarczać do przypisania `REACHABLE`.

Porównywanie wersji musi uwzględniać reguły Maven. Nie można zastąpić ich zwykłym porównaniem tekstowym ani założyć zgodności z SemVer. [Dokumentacja Maven](https://maven.apache.org/pom.html#dependency-version-requirement-specification)

### Obsługa niepewności

- Nierozpoznana wersja oznacza nierozstrzygnięte dopasowanie zakresu.
- Brak metody z mapowania w analizowanej bibliotece daje `UNKNOWN` z wyjaśnieniem, chyba że istnieje osobny dowód rozstrzygający brak podatnego kodu.
- Sprzeczne mapowania są raportowane wraz ze źródłami.
- Osiągalność metody z mapowania częściowego może być raportowana z informacją o jakości dowodu. Brak ścieżki do częściowej listy nie wystarcza do mocnego wyniku negatywnego.
- `NOT_REACHABLE` dotyczy zadeklarowanego modelu analizy i nie jest gwarancją bezpieczeństwa aplikacji.

### Kryterium gotowości

Przykłady z dwiema wersjami biblioteki i przeciążonymi metodami otrzymują właściwe mapowania; brak dopasowania nie zamienia się automatycznie w `NOT_REACHABLE`.

## 3. Analiza kodu przed poprawką

### Cel

Ustalić, jaki kod istniał w podatnej wersji, również gdy poprawka usuwa albo zmienia nazwę metody.

Obecny parser diffów zbiera zmienione linie po stronie kodu po poprawce i pomija czyste usunięcia.

### Nowy przebieg

1. Ariadne identyfikuje commit poprawki i jego rodzica.
2. Pobiera pliki przed zmianą oraz po zmianie.
3. Analizuje obie strony diffu, zachowując stare i nowe ścieżki plików.
4. Ustala metody zmienione, usunięte, dodane lub potencjalnie przemianowane.
5. Sprawdza kandydatów w bytecode analizowanego artefaktu.
6. Zapisuje dowody i stopień rozstrzygnięcia.

GitHub udostępnia informacje o rodzicach commita i zmienionych plikach, które można wykorzystać do odtworzenia obu stron zmiany. [Dokumentacja API GitHub](https://docs.github.com/en/rest/commits/commits#get-a-commit)

Przykład: poprawka usuwa `unsafeParse(String)` i dodaje `safeParse(String, Options)`. Analizując starą bibliotekę, Ariadne powinna szukać `unsafeParse`, zamiast uznawać nową metodę za cel podatności.

### Reguły

- Zmieniona metoda jest kandydatem do mapowania; udział w commicie nie dowodzi podatności.
- Commity dla różnych gałęzi są rozpatrywane oddzielnie.
- Metoda dodana jako zabezpieczenie nie staje się automatycznie metodą podatną.
- Przy merge commitach niejednoznaczny wybór rodzica wymaga oznaczenia niepewności.
- Kod rodzica commita nie jest automatycznie identyczny ze źródłami badanego wydania; zgodność gałęzi, wersji i metod wymaga sprawdzenia.
- Brak źródeł lub błąd pobrania pozostawia nierozstrzygnięty wynik dla danego CVE i widoczny powód.

Pierwszy zakres obejmuje kod Java: modyfikacje, usunięcia i zmiany ścieżek. Wnioskowanie o przemianowaniu metody na podstawie podobieństwa kodu pozostaje jawnie oznaczoną heurystyką.

### Kryterium gotowości

Poprawnie obsłużone są usunięta metoda, zmieniona sygnatura, przeniesiony plik i poprawki dla dwóch gałęzi biblioteki.

## 4. Biblioteki osadzone w shaded JAR-ach

### Cel

Wykryć komponenty umieszczone wewnątrz innego artefaktu i powiązać ich kod z właściwymi CVE.

| Zakres | Zachowanie |
|---|---|
| Biblioteka osadzona bez zmiany nazw pakietów | Rozpoznanie komponentu i przypisanie jego klas w obrębie zewnętrznego JAR-a. |
| Biblioteka po relokacji | Dodatkowo odtworzenie powiązania oryginalnych nazw klas z nazwami w artefakcie. |

Maven Shade może przenosić klasy do innych pakietów i przepisywać odwołania w bytecode. Dopasowanie wyłącznie po oryginalnej nazwie klasy wtedy nie wystarczy. [Dokumentacja Maven Shade](https://maven.apache.org/plugins/maven-shade-plugin/examples/class-relocation.html)

### Dane i dowody

- Zachowane `pom.properties`.
- Dostarczony SBOM.
- Dostępna konfiguracja procesu budowania i relokacji.
- Zgodność zawartości klas z referencyjnym artefaktem.
- W dalszym etapie: odciski kodu odporne na część zmian nazw.

Każdy komponent otrzymuje wersję lub zbiór możliwych wersji oraz podstawę identyfikacji. Sama obecność `pom.properties` nie wystarcza do potwierdzenia, że odpowiadający mu kod pozostał w JAR-ze. SBOM także wymaga powiązania z faktycznym artefaktem i jego zawartością.

### Raport

```text
Komponent: biblioteka X, wersja Y
Lokalizacja: vendor-client.jar
Identyfikacja: potwierdzona / przypuszczalna
Relokacja: original.package → vendor.internal.package
CVE: ...
Mapowanie i osiągalność: ...
```

Jeżeli wersji nie da się rozstrzygnąć, potencjalne CVE pozostają kandydatami z wyjaśnieniem. Nie wybieramy arbitralnie jednej wersji. Przy kilku kopiach klasy konieczne jest ustalenie implementacji wybranej do grafu; same wyniki inwentaryzacji nie mogą tworzyć fikcyjnych ścieżek do pozostałych kopii.

### Kryterium gotowości

Rozpoznanie biblioteki osadzonej bez relokacji, biblioteki z jawną relokacją oraz dwóch kopii w różnych wersjach; niepewne rozpoznanie pozostaje widoczne.

## 5. Świeżość danych i kontrolowane odświeżanie cache

### Cel

Użytkownik wie, z jakich danych powstał wynik, może pobrać aktualne informacje i odtworzyć wcześniejszą analizę.

Obecny cache przechowuje odpowiedzi bez informacji o ich ważności.

| Rodzaj danych | Proponowane zachowanie |
|---|---|
| Odpowiedzi OSV, także brak podatności | Okresowe odświeżanie; początkowo np. co 24 godziny. |
| Szczegóły advisory | Odświeżanie z uwzględnieniem zmian i wycofania rekordu. |
| Kod i diff przypięty pełnym SHA | Długie przechowywanie wraz z hashem zawartości. |
| Wyniki analizy mapowania | Ponowne wyliczenie po zmianie dowodów lub algorytmu. |
| Odpowiedzi LLM | Osobna kontrola odświeżania, z uwagi na koszt i zmienność. |

OSV definiuje m.in. pola `modified` i `withdrawn`; należy zachować je w modelu zamiast traktować advisory jako niezmienny rekord. [Schemat OSV](https://ossf.github.io/osv-schema/)

### Operacje użytkownika

Wszystkie nazwy nowych flag są propozycjami:

- Zwykła analiza używa danych mieszczących się w polityce świeżości.
- `-refresh` wymusza odświeżenie danych podatności.
- `-offline` korzysta wyłącznie z zapisanych danych i pokazuje ich wiek.
- `-max-data-age` pozwala ustawić wymaganie świeżości.
- Analiza na zapisanym snapshotcie odtwarza wcześniejszy stan danych.

Przy nieudanym wymaganym odświeżaniu analiza domyślnie kończy się błędem. Użycie starszych danych wymaga jawnej opcji i oznaczenia raportu jako opartego na nieaktualnych danych. Nie należy automatycznie ponawiać płatnych zapytań LLM przy odświeżaniu danych CVE.

Snapshot obejmuje dokładne odpowiedzi źródeł, mapowania, konfigurację, wersję analizatora i hashe artefaktów. Sam identyfikator bez zachowanej zawartości nie wystarczy. Użytkownik dostarcza oryginalne artefakty lub ich zachowane kopie; system sprawdza zgodność hashy.

### Kryterium gotowości

Ponowne sprawdzenie identycznej aplikacji wykrywa nowo opublikowane CVE; wycofany advisory zostaje prawidłowo obsłużony; analiza offline ujawnia wiek danych; snapshot pozwala odtworzyć wynik.

## 6. Podatności faktycznie używanego JDK/JRE

### Cel i dane wejściowe

Sprawdzać runtime wskazany dla aplikacji z uwzględnieniem producenta, pełnego wydania i poprawek.

Wejściem jest ścieżka do docelowej instalacji, np. proponowane `-jdk-home`, albo plik opisu runtime’u, np. `-runtime-file`. Opis obejmuje producenta/dystrybucję, pełną wersję i build, system operacyjny, architekturę oraz — gdy dotyczy — dokładne wydanie pakietu systemowego.

Ariadne nie utożsamia automatycznie Javy zainstalowanej na komputerze skanującym z runtime’em wdrożonej aplikacji. Wersja bytecode’u określa wymagania zgodności klas, a nie poziom poprawek runtime’u.

### Dwa etapy oceny

1. Identyfikacja podatnego wydania: dopasowanie runtime’u do advisory właściwego producenta lub dystrybucji.
2. Ocena użycia podatnego komponentu: tam, gdzie istnieją mapowania na kod i odpowiedni model analizy, sprawdzenie ścieżek. Podatności JVM, JIT lub kodu natywnego mogą pozostać bez oceny osiągalności metodowej.

Przykładowy wynik:

```text
Wydanie runtime’u objęte advisory: TAK
Osiągalność podatnego kodu: NIEOCENIONA
Powód: podatność w komponencie natywnym JVM
Źródło: advisory producenta
```

Nie przypisujemy wtedy `NOT_REACHABLE`.

Źródła są dobierane do produktu. Ustaleń Oracle nie przenosimy automatycznie na wszystkie dystrybucje OpenJDK. [Przykładowy advisory Oracle](https://www.oracle.com/security-alerts/cpujul2026.html)

Uwzględniamy backporty: poprawka może występować w pakiecie o zachowanym starszym numerze upstream. [Wyjaśnienie Red Hat](https://www.redhat.com/en/blog/what-backporting-and-how-does-it-apply-rhel-and-other-red-hat-products)

Pierwsze wydanie obsługuje jedną jasno określoną dystrybucję. Wybór zależy od docelowych wdrożeń. Przed implementacją adaptera trzeba potwierdzić dostępność i kompletność danych historycznych. Jeśli producent nie podaje pełnych przedziałów wersji, nie wolno wnioskować o nich wyłącznie z listy wersji wymienionych w pojedynczym advisory.

### Kryterium gotowości

Rozróżnienie wydania podatnego i poprawionego, obsługa niepełnego opisu runtime’u oraz jawny brak oceny dla nieobsługiwanego dostawcy.

## 7. Kolejność i pomiar efektów

Najpierw wspólny model wyników i dowodów, następnie świeżość cache i snapshoty, mapowania wersji i sygnatur, analiza obu stron poprawki, komponenty osadzone oraz runtime JDK.

Pierwszy użyteczny zakres obejmuje wspólną podstawę, aktualne i odtwarzalne dane oraz dokładniejsze mapowania z analizą obu stron poprawki. Shaded JAR-y i JDK są kolejnymi przyrostami.

Każdy etap wymaga przypadków kontrolnych z oczekiwanym wynikiem i porównania z obecną wersją Ariadne. Mierzymy poprawność rozpoznania komponentu, mapowania i werdyktu, liczbę błędnych rozstrzygnięć oraz przyczyny `UNKNOWN`. Zmniejszenie liczby `UNKNOWN` jest sukcesem tylko wtedy, gdy nowe rozstrzygnięcia mają sprawdzalne uzasadnienie.

Nie określono czasu realizacji ani procentu pokrycia. Wymagają one pomiarów na reprezentatywnym zestawie aplikacji i przypadków podatności.
