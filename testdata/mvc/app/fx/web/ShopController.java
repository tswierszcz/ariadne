package fx.web;

import fx.trace.Ran;
import org.springframework.ui.Model;
import org.springframework.web.bind.WebDataBinder;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.InitBinder;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ShopController {
    public ShopController() { Ran.stack(); }

    /** Positive: a plain mapping returning a String. */
    @GetMapping("/hello")
    public String hello() { Ran.stack(); return "hello"; }

    /** Positive: a path variable through a registered Converter; the
     * returned object is written by Jackson through its getters. */
    @GetMapping("/item/{id}")
    public Item item(@PathVariable("id") ItemId id) { Ran.stack(); return new Item(id.value(), "lamp"); }

    /** Positive: a JSON body Jackson reads into NewItem (constructor, setter). */
    @PostMapping("/item")
    public String create(@RequestBody NewItem body) { Ran.stack(); return "created " + body.getName(); }

    /** Positive: throws; a @ControllerAdvice @ExceptionHandler answers. */
    @GetMapping("/boom")
    public String boom() { Ran.stack(); throw new BoomException(); }

    /** Unresolved: mapped but never requested on this run — an entry point
     * by design, reached as a root. */
    @GetMapping("/never")
    public String never() { Ran.stack(); return "never"; }

    /** Positive: runs before every handler of this controller. */
    @ModelAttribute
    public void common(Model model) { Ran.stack(); model.addAttribute("shop", "fx"); }

    /** Positive: runs when a handler binds arguments. */
    @InitBinder
    public void binder(WebDataBinder binder) { Ran.stack(); }

    /** Negative: public, unmapped, called by nothing — not an entry point. */
    public String unmapped() { Ran.stack(); return "unmapped"; }
}
