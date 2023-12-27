import psycopg2


def execute_sql(sql_query):
    connection = None
    try:
        connection = psycopg2.connect(
            user="citus",
            password="citus",
            host="127.0.0.1",
            port="5433",
            database="analytics",
        )
        cursor = connection.cursor()
        cursor.execute(sql_query)
        connection.commit()

        if sql_query.lower().startswith("select"):
            records = cursor.fetchall()
            print("Query Results:", records)

    except (Exception, psycopg2.Error) as error:
        print("Error while executing SQL query", error)
    finally:
        if connection:
            cursor.close()
            connection.close()
