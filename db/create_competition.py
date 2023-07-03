import mysql.connector, json, dateutil.parser, os

with open("./competition_data.json") as json_file:
    data = json.load(json_file)

cnx = mysql.connector.connect(
    user=os.environ.get("DB_USER") or "root",
    password=os.environ.get("DB_PASSWORD") or "",
    host=os.environ.get("DB_HOST") or "localhost",
    database=os.environ.get("DB_NAME") or "matemattici_competition",
)

cursor = cnx.cursor()

# create competition
query = (
    "INSERT INTO competitions (id, start_timestamp, end_timestamp) VALUES (%s, %s, %s)"
)
values = (
    data["id"],
    dateutil.parser.isoparse(data["start_timestamp"]).strftime("%Y-%m-%d %H:%M:%S"),
    dateutil.parser.isoparse(data["end_timestamp"]).strftime("%Y-%m-%d %H:%M:%S"),
)
cursor.execute(query, values)

# create problems
query = "INSERT INTO problems (id, competition_id, number, correct_answer) VALUES (%s, %s, %s, %s)"
for problem in data["problems"]:
    cursor.execute(
        query, (problem["id"], data["id"], problem["number"], problem["answer"])
    )

cnx.commit()

cursor.close()
cnx.close()
