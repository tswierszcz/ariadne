package fx.app;

import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

/** A MyBatis mapper: MyBatis implements it with a JDK proxy (MapperProxy). */
@Mapper
public interface PetMapper {
    @Select("select name from pet")
    String name();
}
