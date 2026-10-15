# Callbacki i asynchroniczność — rekomendacja funkcjonalna

Status: plan. Fazy: [ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD](ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD). Zasady: [ROADMAP.MD](ROADMAP.MD).

## Cel i stan obecny

Odtwarzać możliwe wywołania obiektu przekazanego jako zadanie, listener lub funkcja, także gdy ciało wykonuje biblioteka albo inny wątek.

Obecny graf nie ma osobnych modeli rejestracji i wykonania. `DefaultSkip` wyłącza CHA m.in. dla `java.*`; samo rozszerzanie hierarchii nie zapewnia połączenia `Executor.execute` z właściwym `Runnable.run`. Lambda jest obecnie wiązana już przy utworzeniu, co wymaga wspólnej korekty z DYNAMIC_LINKAGE.

## Zachowanie

- Rozróżnić przekazanie/rejestrację callbacka i możliwość jego wykonania w danym API.
- Powiązać konkretne zadanie z konkretnym argumentem i znanym receiverem; nie łączyć wszystkich `run`/`call`/`accept` na classpath.
- Modelować ograniczony zestaw `Thread.start`, `Executor.execute`, wybranych `ExecutorService.submit`, listenerów i operacji CompletionStage.
- Dla API synchronicznego i asynchronicznego pokazać inny rodzaj przejścia. `Executor` jako abstrakcja nie gwarantuje osobnego wątku dla każdej implementacji.
- Dla operatorów reaktywnych oddzielić zbudowanie pipeline od znanego zdarzenia uruchamiającego/subskrypcji; wesprzeć jedną bibliotekę i zakres wersji jako osobny przyrost.
- Zachować warunki uruchomienia, wyjątki i nieznany receiver. Rejestracja, która może się nie powieść, nie jest dowodem wykonania.

### Scenariusze

1. Znany `Runnable` przekazany do `execute` prowadzi do możliwego `run` i dalej do podatnej metody.
2. `Runnable` tylko utworzony i odrzucony nie daje takiej samej ścieżki.
3. Dwa zadania o różnych receiverach nie mieszają przekazywanych danych.
4. Operacja future zależna od sukcesu/porażki ma odpowiedni warunek, a analiza nie udaje, że wszystkie kontynuacje wykonują się bezwarunkowo.

## Raport i granice

Krok w ścieżce pokazuje miejsce rejestracji, callback, model API i warunek. Strzałka asynchroniczna jest relacją możliwego przekazania sterowania, nie stosem wywołań jednego wątku ani dowodem kolejności czasowej.

MVP obejmuje wybrane podstawowe JDK API i znane lokalne zadania. Anulowanie, scheduler policy, wszystkie interleavingi oraz pełna semantyka wszystkich bibliotek reactive nie są objęte pierwszym zakresem. Nieznana możliwość anulowania nie pozwala dowieść, że zadanie nigdy się nie wykona.

Gotowość: modele łączą właściwy obiekt i deskryptor, zachowują warunki oraz nie interpretują każdej lambdy jako wykonanej.

## Źródło

[Executor: kontrakt wykonywania zadań](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/Executor.html). Modele innych API wymagają osobnego potwierdzenia semantyki i wersji przed wdrożeniem.
