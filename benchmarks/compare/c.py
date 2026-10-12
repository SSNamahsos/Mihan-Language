d = {}
for i in range(300000):
    k = "k" + str(i % 1000)
    d[k] = d.get(k, 0) + i
xs = []
for i in range(300000):
    xs.append((i * 7919) % 100003)
print(len(d), sorted(xs)[1000])
