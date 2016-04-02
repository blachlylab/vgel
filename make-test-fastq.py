import random

with open('test.fastq','w') as fo:
    for i in range(100000):
        rand_len = int(random.gauss(50, 10))
        
        fo.write("@JSB:1:FC000:1:{}:{}:{} 0:N:0:ACTG\n".format(str(rand_len), str(i), str(i)))
        fo.write(("A" * rand_len) + "\n")
        fo.write("+\n")
        fo.write(("@" * rand_len) + "\n")

