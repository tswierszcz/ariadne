package fx.web;

import fx.trace.Ran;
import org.springframework.core.convert.converter.Converter;

/** Registered through WebMvcConfigurer#addFormatters. */
public class IdConverter implements Converter<String, ItemId> {
    public IdConverter() { Ran.stack(); }

    @Override
    public ItemId convert(String source) { Ran.stack(); return new ItemId(source); }
}
