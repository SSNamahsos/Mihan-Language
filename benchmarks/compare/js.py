import json
rows = []
for i in range(50000):
    rows.append({"id": i, "nam": "kar" + str(i), "dorost": i % 2 == 0, "vazn": i * 0.5})
s = json.dumps(rows)
back = json.loads(s)
print(len(s), len(back))
