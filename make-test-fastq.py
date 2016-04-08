import random

def writer(read_count, randomly_generated_lengths):
    '''
    Creates a dummy fastq file for testing 
    
    Input: desired read count (int), whether read lengths should be random (bool)
    Output: writes a fastq file to 'test.fastq'
    '''
    with open('test.fastq','w') as fo:
        for i in range(read_count):
            if bool(randomly_generated_lengths):
                rand_len = int(random.gauss(50, 10))
            else:
                rand_len = 50
            if rand_len < 0: rand_len = (rand_len * -1) # desireable not to have zero length seq at least for testing
            elif rand_len == 0: rand_len = 1    # desireable not to have zero length seq at least for testing 
            fo.write("@JSB:1:FC000:1:{}:{}:{} 0:N:0:ACTG\n".format(str(rand_len), str(i), str(i)))
            fo.write(("A" * rand_len) + "\n")
            fo.write("+\n")
            fo.write(("@" * rand_len) + "\n")

def main():
    # canonical usage
    writer(100000, True)
    
    # write an example that will hit the barchart index panic 
    #writer(2, False)

if __name__ == "__main__":
    main()