import psycopg2


def check_table_schema(table_name):
    try:
        connection = psycopg2.connect(
            user="your_username",
            password="your_password",
            host="127.0.0.1",
            port="5432",
            database="your_database",
        )
        cursor = connection.cursor()
        query = f"SELECT table_schema FROM information_schema.tables WHERE table_name = '{table_name}'"
        cursor.execute(query)
        schema_name = cursor.fetchone()
        return schema_name[0] if schema_name else None
    except (Exception, psycopg2.Error) as error:
        print("Error while checking table schema", error)
        return None
    finally:
        if connection:
            cursor.close()
            connection.close()
