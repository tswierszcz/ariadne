# MethodHandle i dynamiczne wiązanie — rekomendacja funkcjonalna

Status: plan. Fazy: [DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD](DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD). Wspólne wartości: [DATA_FLOW.MD](DATA_FLOW.MD).

## Cel i punkt wyjścia

Rozpoznawać cele uchwytów metod i wybranych mechanizmów bootstrap, nie wykonując ich podczas skanowania. Obecny `resolveInvokeDynamic` obsługuje LambdaMetafactory i zwraca jeden `CallDynamic`; informacja o rodzaju implementacyjnego MethodHandle jest dostępna w parserze, ale nie kieruje pełnym modelem dispatch w grafie.

Obecne przypięcie ciała lambdy do miejsca jej utworzenia jest uproszczeniem. Docelowo utworzenie obiektu funkcyjnego, jego użycie i ewentualny wirtualny dispatch muszą być rozróżnione.

## Funkcje

| Mechanizm | Zachowanie docelowe |
|---|---|
| Lambda i method reference | Powiązać obiekt funkcyjny z celem, receiverem i miejscem utworzenia; modelować wywołanie przy użyciu. |
| Stały MethodHandle | Zachować rodzaj, klasę, nazwę i typ celu; samo posiadanie uchwytu nie jest wykonaniem jego ciała. |
| Lookup findStatic/findVirtual/findConstructor | Rozwiązać znany cel zgodnie z danymi i ograniczeniami dostępu. |
| invoke/invokeExact | Zastosować odpowiednią semantykę typu i dispatch przy znanych wartościach. |
| Adaptacje uchwytów | Wdrażać wybrane `bindTo`, `asType` i kombinatory z jawną listą pokrycia. |
| Inne invokedynamic / CONSTANT_Dynamic | Zidentyfikować bootstrap i obsłużyć znany model albo zgłosić nierozstrzygnięte wiązanie. |

## Scenariusze i wynik

- Niewykorzystana lambda nie powinna sama dowodzić wywołania jej ciała.
- Method reference do metody wirtualnej zachowuje możliwe implementacje receivera.
- Stały uchwyt prowadzi do celu dopiero przy invocation; niezgodny `invokeExact` nie jest automatycznie poprawnym wywołaniem.
- Nieznany custom bootstrap daje lukę, a jeśli znana jest metoda bootstrap, można osobno pokazać możliwość wykonania jej kodu podczas wiązania.
- Konkatenacja tekstu nie może być bezwarunkowo uznana za semantycznie pustą: konwersja obiektu może wywołać kod, którego model wymaga osobnej reguły.

Raport przedstawia bootstrap, rodzaj uchwytu, cel, sposób wywołania, warunki i nieobsługiwane adaptacje. Sam model dynamicznej krawędzi nie potwierdza exploitability.

## Granice i gotowość

MVP: poprawna semantyka tworzonych/używanych lambd i ograniczony katalog stałych uchwytów. Potem wybrane Lookup i adaptacje. Arbitralny bootstrap i MutableCallSite zależny od runtime’u pozostają częściowe; służą temu ślady z RUNTIME_NATIVE.

Gotowość: rodzaj dispatch nie ginie, niewykorzystane obiekty funkcyjne nie tworzą pozornej obserwacji wykonania, a nieznany bootstrap i błędny typ mają osobne powody.

## Źródła

- [LambdaMetafactory](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/lang/invoke/LambdaMetafactory.html).
- [MethodHandles.Lookup](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/lang/invoke/MethodHandles.Lookup.html).
- [JVM: dynamiczne stałe i miejsca wywołań](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-5.html#jvms-5.4.3.6).
