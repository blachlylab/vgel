import random

with open('test.fastq','w') as fo:
    for i in range(1000000):
        rand_len = int(random.gauss(50, 10))
        if rand_len < 0: rand_len = (rand_len * -1) # desireable not to have zero length seq at least for testing
        elif rand_len == 0: rand_len = 1    # desireable not to have zero length seq at least for testing 
        fo.write("@JSB:1:FC000:1:{}:{}:{} 0:N:0:ACTG\n".format(str(rand_len), str(i), str(i)))
        fo.write(("A" * rand_len) + "\n")
        fo.write("+\n")
        fo.write(("@" * rand_len) + "\n")

