import mysql.connector

cnx = mysql.connector.connect(
    user="root",
    password="",
    host="localhost",
    database="matemattici_competition",
)

cursor = cnx.cursor()

cursor.execute("DELETE FROM submissions")
cursor.execute("DELETE FROM problems")
cursor.execute("DELETE FROM competitions")

cnx.commit()

# create competition
query = "INSERT INTO competitions (id, start_timestamp, end_timestamp) VALUES (%s, CURRENT_TIMESTAMP(), DATE_ADD(CURRENT_TIMESTAMP(), INTERVAL 1 HOUR))"
values = (1234,)
cursor.execute(query, values)

# create problems
query = "INSERT INTO problems (id, competition_id, number, correct_answer) VALUES (%s, 1234, %s, %s)"
for i in range(1, 6):
    cursor.execute(query, (i, i, i))

cnx.commit()

cursor.close()
cnx.close()
