/*Archivo de Configuraciones que necesitaremos cuando generemos el Docker
Por ejemplo crear la BBDD*/

SET @MYSQLDUMP_TEMP_LOG_BIN = @@SESSION.SQL_LOG_BIN;
SET @@SESSION.SQL_LOG_BIN= 0;

SET @@GLOBAL.GTID_PURGED=/*!80000 '+'*/ '';

/*Crea la base de datos go_course_web si no existe.*/
CREATE DATABASE IF NOT EXISTS `go_micro_user`;