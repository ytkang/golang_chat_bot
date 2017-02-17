import csv
import time
import datetime
from pymongo import MongoClient
from bson.objectid import ObjectId

mongoClient = MongoClient(host="127.0.0.1", port=27017, maxPoolSize=50, waitQueueMultiple=10)
db = mongoClient["chat"]

prev_time = 0
pre_user = None
long_interval = 600
prev_id = None

with open('data.csv', newline='', encoding='utf-8') as f:
    reader = csv.reader(f)
    for row in reader:
        timestring = row[0]
        user = row[1]
        message = row[2]
        timestamp = time.mktime(datetime.datetime.strptime(timestring, "%Y-%m-%d %H:%M:%S").timetuple())
        prev_time = timestamp

        if len(message) < 2 or pre_user == user:
            continue

        message = message.replace('(Emoticon)', '')
        if timestamp > prev_time + long_interval:
            print("insert by time")
            doc = db.messages.update_one({'text': message}, {'$set': {'text': message}}, upsert=True)
            prev_id = doc.upserted_id
        else:
            if prev_id is not None:
                print("update_one")
                db.messages.update_one({'text': message}, {'$set': {'text': message}, '$addToSet': {'answerOf': ObjectId(prev_id)}}, upsert=True)
                prev_id = None
            else:
                print("insert")
                doc = db.messages.find_and_modify({'text': message}, {'text': message}, upsert=True, new=True)
                prev_id = doc['_id']

        pre_user = user