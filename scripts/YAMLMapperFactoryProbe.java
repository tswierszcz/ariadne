import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.fasterxml.jackson.dataformat.yaml.YAMLMapper;

/**
 * JVM check behind the 2026-11-18 correction of LIBRARY_MODE lb2-4's
 * leftovers: through YAMLMapper's public API the inherited
 * ObjectMapper#createParser does reach the JSON parser. YAMLMapper(YAMLFactory)
 * accepts null, and ObjectMapper(JsonFactory) then stores a
 * MappingJsonFactory in _jsonFactory — so the path
 * ObjectMapper#createParser -> _jsonFactory.createParser -> JSON parser is
 * real, not an over-approximation of the field receiver.
 *
 * Usage (jars from scripts/library-measure.sh):
 *   java -cp "$(ls work/libs/jackson-yaml/lib.jar work/libs/jackson-yaml/deps/*.jar | tr '\n' ':')" scripts/YAMLMapperFactoryProbe.java
 * Output with jackson-dataformat-yaml 2.12.3 / databind 2.12.3:
 *   YAMLMapper()     parser -> com.fasterxml.jackson.dataformat.yaml.YAMLParser
 *   YAMLMapper(null) parser -> com.fasterxml.jackson.core.json.ReaderBasedJsonParser
 */
public class YAMLMapperFactoryProbe {
    public static void main(String[] a) throws Exception {
        System.out.println("YAMLMapper()     parser -> " + new YAMLMapper().createParser("a: 1").getClass().getName());
        System.out.println("YAMLMapper(null) parser -> " + new YAMLMapper((YAMLFactory) null).createParser("{\"a\":1}").getClass().getName());
    }
}
