package fx.web;

import fx.trace.Ran;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestControllerAdvice;

@RestControllerAdvice
public class Errors {
    @ExceptionHandler(BoomException.class)
    @ResponseStatus(HttpStatus.CONFLICT)
    public String onBoom(BoomException e) { Ran.stack(); return "handled"; }
}
