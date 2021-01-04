import os
DIR = "/Users/chenfengjin/baidu/xuperchain/core/contractsdk/go/example/tests"


def parseFile(filename):
    output = []
    with open(os.path.join(DIR, filename)) as f:
        lines = f.readlines()
        for line in lines:
            if line.startswith("var lang"):
                line = "var lang=\"java\""+"\n"
            if line.startswith("var codePath"):
                # print(line)
                line = "var codePath=\"../"+filename.split(".")[0]+"/target/"+filename.split(".")[
                    0]+"-0.1.0-jar-with-dependencies.jar\"\n"
            if line.startswith("var type="):
                line = "var type=\"native\""+"\n"
            if "c.Invoke" in line:
                prefix = "var resp = c.Invoke("
                suffix = ",".join(line.split(",")[1:])
                method = line.split(",")[0].split("(")[1].strip("\"")
                method = method[:1].lower() + method[1:]
                line = prefix+"\""+method+"\","+suffix
            output.append(line)
        with open("test/"+filename, "w") as out:
            out.writelines(output)


def main():
    files = [i for i in os.listdir(DIR) if i.endswith(".test.js")]
    for file in files:
        parseFile(file)


if __name__ == "__main__":
    main()
