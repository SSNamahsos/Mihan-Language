class Hesab:
    def __init__(self):
        self.mojudi = 0
    def variz(self, m):
        self.mojudi += m

h = Hesab()
for i in range(1000000):
    h.variz(i % 7)
dobar = lambda x: x * 2
s = 0
for i in range(300000):
    s += dobar(i)
print(h.mojudi, s)
