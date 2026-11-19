package fx;

import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface UserMapper {
    @Select("SELECT COUNT(*) FROM INFORMATION_SCHEMA.SYSTEM_USERS")
    int users();
}
