# Refleksja — fazy R0–R4

Status: plan. Wymagania: [REFLECTION.MD](REFLECTION.MD). Zasady B0–B4: [ROADMAP.MD](ROADMAP.MD).

## R0. Zakres API i obiekty refleksji

Zależności: JVM G0–G3, DATA_FLOW D0–D2, B0–B3.

- Zaprojektować modele `ClassValue`, `ReflectiveMethod`, `ReflectiveConstructor`, receiver i zestaw parametrów w wspólnej domenie wartości.
- Wprowadzić rejestr obsługiwanych sygnatur i różnic wersji. `Class.forName` i `ClassLoader.loadClass` nie są jednym modelem.
- Zdefiniować powody: nieznana nazwa, nieznany loader, brak członka, niepełne parametry, nieustalony dostęp oraz przekroczony budżet.

Gotowe: modele można uruchomić przy miejscu wywołania bez wykonywania analizowanego kodu.

## R1. Stałe i lokalne wyszukiwanie

- Odczytać wartości argumentów z DATA_FLOW D2; nie tworzyć osobnego heurystycznego skanera „poprzedniego ldc”.
- Implementować `forName` dla jednoznacznej nazwy, literały klas i bazowy lookup publicznych metod/konstruktorów w znanym kontekście.
- Rozróżnić `getMethod` i `getDeclaredMethod` oraz listę typów parametrów. Nie dopasowywać wszystkich przeciążeń jako dokładnego wyniku.
- Zachować wynik wyszukania jako obiekt i informacje o możliwych wyjątkach. Używać JVM G3 dla reguł inicjalizacji.

Gotowe: literalna nazwa i nazwa przez zmienną dają ten sam wynik; samo wyszukanie nie dodaje wywołania metody.

## R2. Wywołanie i alokacje

- Dla `Method.invoke` odtworzyć cel oraz receiver; uwzględnić rzeczywisty dispatch zamiast bezwarunkowego przypięcia do dowolnej nadklasy.
- Dla `Constructor.newInstance` modelować alokację, konstruktor i związane inicjalizacje; przekazać typ do rejestru B3/RTA.
- Sprawdzać zgodność liczby/typów argumentów w zakresie obsługiwanej semantyki. Nie zakładać skutecznego invocation przy dowiedzionej niezgodności.
- Dodać dowody krawędzi i widok unresolved sites; rozsądny konserwatywny fallback dla nieznanego receivera pozostaje oznaczony.

Gotowe: pozytywna ścieżka, dwa obiekty Method, overload i błędne argumenty mają odrębne wyniki. R0–R2 stanowią pierwsze wydanie.

## R3. Przepływ między metodami, konfiguracja i dostęp

Zależności: DATA_FLOW D3, FRAMEWORK F1 dla scenariuszy konfiguracji; MODULE dane SERVICE S3 tam, gdzie potrzebne.

- Przenosić nazwy i obiekty refleksji przez fabryki i argumenty. Oddzielić dwa wywołania tej samej fabryki w zakresie przyjętej precyzji.
- Przyjmować wartości z jawną konfiguracją i hashem źródła. Nie pobierać automatycznie środowiska/sekretów hosta jako konfiguracji aplikacji.
- Wdrożyć wybrany zakres `getDeclared*`, access checks i `setAccessible`/`trySetAccessible` z uwzględnieniem docelowej Javy i modułów. Nieznane uprawnienia traktować jako warunek.
- Dla skończonych zbiorów celów respektować limity; po poszerzeniu zachować lukę i uzasadnienie zamiast uciętych „pewnych” celów.

Gotowe: aktywny profil, dwie nazwy i odmowa dostępu nie są mylone ze stałą jednego celu.

## R4. Kwalifikacja i integracja

- Fixtures obejmują inicjalizację true/false, samo wyszukanie, invoke, konstruktor, dziedziczenie, overload, factory, nieznany parametr i niedostępną metodę.
- W `internal/reach` i raporcie przenieść wpływ unresolved sites zgodnie z B0; nie globalnie wyciszać wszystkich CVE ani udawać, że luka jest nieistotna.
- Wersjonować modele i dołączyć ich wejścia do snapshotów CVE F3.
- Wykonać kontrole B4 i udokumentować pokrycie według API/wersji; pomiary jakości przed rozszerzaniem heurystyk.

Poza pierwszym zakresem: dowolne runtime-computed nazwy, pełna logika security/loaderów oraz utożsamianie obserwacji z deklaracją. MethodHandle ma własny plan DYNAMIC_LINKAGE, korzystający z tych samych wartości.
