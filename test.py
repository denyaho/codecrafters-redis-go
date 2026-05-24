

pystr = str(input())
N = len(pystr)
duplist = {}

len_substr = 1e9

for i in range(N):
    if pystr[i] not in duplist:
        duplist[pystr[i]] = i
    else:
        len_substr = min(len_substr, i - duplist[pystr[i]])

if len_substr == 1e9:
    len_substr = N
print(len_substr)