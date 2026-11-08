package cg;

// Interface with implementations at several depths, one abstract.
interface Task {
    void run();
}

class TaskA implements Task {
    public void run() { Trace.hit(); }
}

class TaskB implements Task {
    public void run() { Trace.hit(); }
}

abstract class AbstractTask implements Task {
}

class TaskC extends AbstractTask {
    public void run() { Trace.hit(); }
}

// Overloads: a subclass overrides only the String overload.
class Over {
    void f(int x) { Trace.hit(); }
    void f(String s) { Trace.hit(); }
}

class OverSub extends Over {
    @Override
    void f(String s) { Trace.hit(); }
}

// Default method: one implementation keeps it, one overrides it.
interface Greeter {
    default String hello() { Trace.hit(); return "default"; }
}

class PlainGreeter implements Greeter {
}

class LoudGreeter implements Greeter {
    public String hello() { Trace.hit(); return "LOUD"; }
}

// Generic bridge: StrFn gets a compiler bridge apply(Object)Object.
interface Fn<T> {
    T apply(T t);
}

class StrFn implements Fn<String> {
    public String apply(String s) { Trace.hit(); return s; }
}

// Covariant return: Dog gets a bridge self()Animal.
class Animal {
    Animal self() { Trace.hit(); return this; }
}

class Dog extends Animal {
    @Override
    Dog self() { Trace.hit(); return this; }
}

// Inherited, not overridden: the call names Sub, Base declares the body.
class Base {
    void hook() { Trace.hit(); }
}

class Sub extends Base {
}

// super call: invokespecial, exactly the parent's body.
class Parent {
    void m() { Trace.hit(); }
}

class Child extends Parent {
    @Override
    void m() { Trace.hit(); }

    void callSuper() { super.m(); }
}
