package fx.beans;

/** Made by WidgetFactory (a FactoryBean), never by a constructor call in the input but its own. */
public class Widget {
    final String label;
    Widget(String label) { this.label = label; }
}
