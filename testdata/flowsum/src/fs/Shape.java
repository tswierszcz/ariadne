package fs;

abstract class Shape {
    abstract String name(String s);
}

class Circle extends Shape {
    String name(String s) {
        return s;
    }
}

class Square extends Shape {
    String name(String s) {
        return "square";
    }
}
