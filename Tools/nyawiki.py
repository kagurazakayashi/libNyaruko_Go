#coding: utf-8
#!/usr/bin/python3
# Wiki 生成器

import sys
import opencc

if len(sys.argv) <= 1:
    exit(1)

f = open(sys.argv[1], 'r', encoding="utf-8")
result = list()
mode = -1
tmpStr = ""
print("# " + sys.argv[1].strip())
opencconv = opencc.OpenCC('tw2sp.json')
for line in f.readlines():
    line = line.strip()
    line = opencconv.convert(line)
    if mode == -1:
        tmpStr = "**" + line.replace("//", "").strip() + "**\n\n"
        mode = 0
    elif mode == 0:
        if "//" in line and ": " in line:
            tmpStr += "## " + line.split(": ")[1] + "\n- `<[code]>`\n"
            mode = 2
    elif mode == 2 and "func " not in line:
        line = line.replace("//", "")
        tcount =  line.count("\t", 0)
        line = line.replace("\t", "")
        tmpStr += (" " * tcount) + "- " + line + "\n"
    if mode > 0 and "func " in line:
        tmpStr = tmpStr.replace("<[code]>", line.replace(" {", ""))
        print(tmpStr)
        tmpStr = ""
        mode = 0
f.close()
