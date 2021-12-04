import os
import numpy as np

from datetime import datetime
################################ FUNCTIONS WE ARE GIVEN ############################################
####################################################################################################
def read_fasta_file(filename):
    """
    Reads the given FASTA file f and returns a dictionary of sequences.

    Lines starting with ';' in the FASTA file are ignored.
    """
    sequences_lines = {}
    current_sequence_lines = None
    with open(filename) as fp:
        for line in fp:
            line = line.strip()
            if line.startswith(';') or not line:
                continue
            if line.startswith('>'):
                sequence_name = line.lstrip('>')
                current_sequence_lines = []
                sequences_lines[sequence_name] = current_sequence_lines
            else:
                if current_sequence_lines is not None:
                    current_sequence_lines.append(line)
    sequences = {}
    for name, lines in sequences_lines.items():
        sequences[name] = ''.join(lines)
    return sequences


def translate_observations_to_indices(obs):
    mapping = {'a': 0, 'c': 1, 'g': 2, 't': 3}
    return [mapping[symbol.lower()] for symbol in obs]

def translate_states_to_meta_states(z):
    return "".join(["N" if zi == "3" else ("C" if (zi == "0") or (zi == "1") or (zi == "2") else "R") for zi in z])

def translate_meta_states_to_states(ann):
    list = []
    last_a = ""
    for a in ann:
        toAppend = ""
        if a == "N":
            toAppend = "3"
        elif a == "C":
            if last_a == "3":
                toAppend = "2"
            elif last_a == "2":
                toAppend = "1"
            elif last_a == "1":
                toAppend = "0"
            elif last_a == "0":
                toAppend = "2"
            elif last_a == "6":
                toAppend = "2"
        elif a == "R":
            if last_a == "3":
                toAppend ="4"
            elif last_a == "4":
                toAppend ="5"
            elif last_a == "5":
                toAppend = "6"
            elif last_a == "6":
                toAppend = "4"      
            elif last_a == "0":
                toAppend = "4"
        last_a = toAppend
        list.append(toAppend)
    return ''.join(list)
    

def translate_indices_to_observations(indices):
    mapping = ['a', 'c', 'g', 't']
    return ''.join(mapping[idx] for idx in indices)

def translate_path_to_indices(path):
    return list(map(lambda x: int(x), path))

def translate_indices_to_path(indices):
    return ''.join([str(i) for i in indices])

def log(x):
        if x == 0:
            return float('-inf')
        return np.log(x)

def nlog(x):
    return np.vectorize(log)(x)

class hmm:
    def compute_w_log_opt(self, X): 
        init_probs = nlog(self.init_probs)
        emission_probs = nlog(self.emission_probs)
        trans_probs = nlog(self.trans_probs)

        x = translate_observations_to_indices(X)
        K = len(self.init_probs)
        n = len(x)

        allTransitions = [list() for x in range(K)]
        for col in range(K):
            transitions = []
            for row in range(K):
                if trans_probs[row][col] != float('-inf'):
                    transitions.append(row)
            print("These are the transistions: ", transitions)         
            allTransitions[col].extend(transitions)
        print("Alltransitions:", allTransitions)


        w = np.zeros((K, n))
        # BASE CASE: fill out w[i][0] for i = 0..k-1
        for i in range(K): 
            w[i,0] = init_probs[i] + emission_probs[i][x[0]] #Prob of state z_i * prob x[0] given z_i

        # Inductive case: fill out w[i][j] for i = 0..k, j = 0..n-1
        for col in range(1,n): #For each column in omega
            for row in range(K): #For each row (each z-state) in omega
                max_so_far = float('-inf')
                #max = np.max(w[:][col-1] * model.trans_probs[:][row])
                for z_i in allTransitions[row]: #For each row in previous column
                    prev_prob = w[z_i,col-1] + trans_probs[z_i][row]
                    if prev_prob > max_so_far: 
                        max_so_far = prev_prob
                w[row,col] = emission_probs[row][x[col]] + max_so_far
        print("Computed w")
        return w 

    def backtrack_log_opt(self, X, w):
        init_probs = nlog(self.init_probs)
        emission_probs = nlog(self.emission_probs)
        trans_probs = nlog(self.trans_probs)

        pathList = []
        n = w.shape[1]
        x = translate_observations_to_indices(X)
        K = len(init_probs)

        allTransitions = [list() for x in range(K)]
        for col in range(K):
            transitions = []
            for row in range(K):
                if trans_probs[row][col] != float('-inf'):
                    transitions.append(row)
            print("These are the transistions: ", transitions)         
            allTransitions[col].extend(transitions)
        print("Alltransitions:", allTransitions)

        #n-1
        lastZ = np.argmax(w[:, n-1])
        lastZProb = np.max(w[:, n-1])
        pathList.insert(0, lastZ)

        #n-2 to 1
        prev = lastZ
        prevProb = lastZProb
        for col in range(n-2, -1, -1):
            for row in range(K): #For each row (each z-state) in omega
                max_so_far = float('-inf')
                for z_i in allTransitions[row]: #For each row in previous column
                    prev_prob = w[z_i,col-1] + trans_probs[z_i][row]
                    if prev_prob > max_so_far: 
                        max_so_far = prev_prob
                w[row,col] = emission_probs[row][x[col]] + max_so_far

        #The original:
        prev = lastZ
        prevProb = lastZProb
        for col in range(n-2, -1, -1):
            columnw = w[:, col]
            emissionProb = emission_probs[prev][x[col+1]]
            for index, val in enumerate(columnw): 
                currentProb = emissionProb + (val + trans_probs[index][prev])
                if prevProb == currentProb:
                    prevProb = val
                    prev = index
                    pathList.insert(0,prev)
                    break

        realPath = translate_indices_to_path(pathList)        
        return realPath


    def __init__(self, init_probs, trans_probs, emission_probs):
        self.init_probs = init_probs
        self.trans_probs = trans_probs
        self.emission_probs = emission_probs
        self.K = init_probs.shape[0]
        self.D = emission_probs.shape[1]

    def count_transitions_and_emissions(self, x, z):
        """
        Returns a KxK matrix and a KxD matrix containing counts cf. above
        """
        K = self.K
        D = self.D
        A_count = np.zeros((K,K))
        phi_count = np.zeros((K,D))
        ACindicies = translate_observations_to_indices(x)
        pathList = translate_path_to_indices(z)
        for i in range(len(pathList)-1):
            ival = pathList[i]
            jval = pathList[i+1]
            A_count[ival,jval] += 1
            phi_count[ival, ACindicies[i]] += 1
        phi_count[pathList[len(pathList)-1], ACindicies[len(pathList)-1]] += 1
        return A_count, phi_count
        
    def training_by_counting(self, A, fi):
        """
        Returns a HMM trained on x and z cf. training-by-counting.
        """
        for i in range(A.shape[0]):
            sum = np.sum(A[i,:])
            if sum == 0:
                A[i,:] = 0
            else:
                A[i,:] /= sum
            sum2  = np.sum(fi[i,:])
            if sum2 == 0: 
                fi[i,:] = 0 
            else:
                fi[i,:] /= sum2
        self.trans_probs = A
        self.emission_probs = fi
        print("Done training")

    def training_by_counting_full(self, x, z):
            A, fi = self.count_transitions_and_emissions(x, z)
            self.training_by_counting(A, fi)

    def forward(self, xs):
        n = len(xs)
        self.omega = np.zeros([self.K, n])
    
    def compute_w_log(self, X): 
        init_probs = nlog(self.init_probs)
        emission_probs = nlog(self.emission_probs)
        trans_probs = nlog(self.trans_probs)

        x = translate_observations_to_indices(X)
        K = len(self.init_probs)
        n = len(x)

        w = np.zeros((K, n))
        # BASE CASE: fill out w[i][0] for i = 0..k-1
        for i in range(K): 
            w[i,0] = init_probs[i] + emission_probs[i][x[0]] #Prob of state z_i * prob x[0] given z_i

        # Inductive case: fill out w[i][j] for i = 0..k, j = 0..n-1
        for col in range(1,n): #For each column in omega
            for row in range(K): #For each row (each z-state) in omega
                max_so_far = float('-inf')
                #max = np.max(w[:][col-1] * model.trans_probs[:][row])
                for z_i in range(K): #For each row in previous column
                    prev_prob = w[z_i,col-1] + trans_probs[z_i][row]
                    if prev_prob > max_so_far: 
                        max_so_far = prev_prob
                w[row,col] = emission_probs[row][x[col]] + max_so_far
        print("Computed w")
        return w 

    def opt_path_prob_log(self, w):
        n = w.shape[1]
        probForLastZ = np.max(w[:, n-1]) #... I gave you my heart, but the very next day you gave it awayyyyyy.... Thiiiiiis year...
        return probForLastZ
    
    def backtrack_log(self, X, w):
        #init_probs = nlog(model.init_probs)
        emission_probs = nlog(self.emission_probs)
        trans_probs = nlog(self.trans_probs)

        pathList = []
        n = w.shape[1]
        x = translate_observations_to_indices(X)

        #n-1
        lastZ = np.argmax(w[:, n-1])
        lastZProb = np.max(w[:, n-1])
        pathList.insert(0, lastZ)

        #n-2 to 1
        prev = lastZ
        prevProb = lastZProb
        for col in range(n-2, -1, -1):
            columnw = w[:, col]
            emissionProb = emission_probs[prev][x[col+1]]
            for index, val in enumerate(columnw): 
                currentProb = emissionProb + (val + trans_probs[index][prev])
                if prevProb == currentProb:
                    prevProb = val
                    prev = index
                    pathList.insert(0,prev)
                    break

        realPath = translate_indices_to_path(pathList)        
        return realPath

    def viterbi_update_model(self, x):
        """
        return a new model that corresponds to one round of Viterbi training, 
        i.e. a model where the parameters reflect training by counting on x 
        and z_vit, where z_vit is the Viterbi decoding of x under the given 
        model.
        """
        
        # Your code here ...
        w = self.compute_w_log(x)
        z_vit = self.backtrack_log(x, w)
        
        K = len(self.trans_probs)
        D = len(self.emission_probs[0])
        print("K: ", K)
        print("D: ", D)
        trans, emission = self.training_by_counting(x, z_vit)

        
        return trans, emission
    
    def viterbi_training(self, x):
        trans, emission = self.viterbi_update_model(x)
        model_unchanged = np.allclose(trans, self.trans_probs) and np.allclose(emission, self.emission_probs)
        print("Oldtrans:", self.trans_probs)
        print("Newtrans: ", trans)
        if model_unchanged:
            return trans, emission
        return trans,emission

def makeHmm():
    init_probs_7_state = np.array([0.00, 0.00, 0.00, 1.00, 0.00, 0.00, 0.00])

    trans_probs_7_state = np.array([
        [0.00, 0.00, 0.90, 0.10, 0.00, 0.00, 0.00],
        [1.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00],
        [0.00, 1.00, 0.00, 0.00, 0.00, 0.00, 0.00],
        [0.00, 0.00, 0.05, 0.90, 0.05, 0.00, 0.00],
        [0.00, 0.00, 0.00, 0.00, 0.00, 1.00, 0.00],
        [0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 1.00],
        [0.00, 0.00, 0.00, 0.10, 0.90, 0.00, 0.00],
    ])

    emission_probs_7_state = np.array([
        #   A     C     G     T
        [0.30, 0.25, 0.25, 0.20],
        [0.20, 0.35, 0.15, 0.30],
        [0.40, 0.15, 0.20, 0.25],
        [0.25, 0.25, 0.25, 0.25],
        [0.20, 0.40, 0.30, 0.10],
        [0.30, 0.20, 0.30, 0.20],
        [0.15, 0.30, 0.20, 0.35],
    ])

    hmm_7_state = hmm(init_probs_7_state, trans_probs_7_state, emission_probs_7_state)
    return hmm_7_state
    
def validate_hmm(model):
    pi = np.array( model.init_probs)
    A = np.array(model.trans_probs)
    fi = np.array(model.emission_probs)
    for i in range(A.shape[0]):
        rowSum = 0
        for j in range(A.shape[1]):
            rowSum += A[i,j]
        if not(np.allclose(rowSum, 1)):
            return False
    sum = 0
    for i in range(pi.shape[0]):
        sum += pi[i]
    if not(np.allclose(sum, 1)):
        print("Oh no")
    for i in range(fi.shape[0]):
        rowSum = 0
        for j in range(fi.shape[1]):
            rowSum += fi[i,j]
        if not(np.allclose(rowSum, 1)):
            return False
    return True

################################ FUNCTIONS MADE BY US ##############################################
####################################################################################################
def read_fasta_file_as_string(name, amount):
    toLoad = name + ".fa"
    
    fastaFile = read_fasta_file("ASSIGNMENTS2021\Machine Learning\Assignment 3\FastaFiles\\" + toLoad)
    genome = fastaFile[name]
    if amount == 0:
        return genome
    else:
        return genome[:amount]

def compute_accuracy(true_ann, pred_ann):
    if len(true_ann) != len(pred_ann):
        print("Lengths are NOT the same..")
        print("These are the lengths:")
        print("len true:", len(true_ann))
        print("len pred:", len(pred_ann))
        return 0.0
    return sum(1 if true_ann[i] == pred_ann[i] else 0 
               for i in range(len(true_ann))) / len(true_ann)


if __name__ == '__main__':
    #make (and validate) the model
    hmm = makeHmm()
    valid = validate_hmm(hmm)
    if not valid:
        print("Model is not valid - something is wrong!!!")
        quit()

    #read some data
    x_long = 'TGAGTATCACTTAGGTCTATGTCTAGTCGTCTTTCGTAATGTTTGGTCTTGTCACCAGTTATCCTATGGCGCTCCGAGTCTGGTTCTCGAAATAAGCATCCCCGCCCAAGTCATGCACCCGTTTGTGTTCTTCGCCGACTTGAGCGACTTAATGAGGATGCCACTCGTCACCATCTTGAACATGCCACCAACGAGGTTGCCGCCGTCCATTATAACTACAACCTAGACAATTTTCGCTTTAGGTCCATTCACTAGGCCGAAATCCGCTGGAGTAAGCACAAAGCTCGTATAGGCAAAACCGACTCCATGAGTCTGCCTCCCGACCATTCCCATCAAAATACGCTATCAATACTAAAAAAATGACGGTTCAGCCTCACCCGGATGCTCGAGACAGCACACGGACATGATAGCGAACGTGACCAGTGTAGTGGCCCAGGGGAACCGCCGCGCCATTTTGTTCATGGCCCCGCTGCCGAATATTTCGATCCCAGCTAGAGTAATGACCTGTAGCTTAAACCCACTTTTGGCCCAAACTAGAGCAACAATCGGAATGGCTGAAGTGAATGCCGGCATGCCCTCAGCTCTAAGCGCCTCGATCGCAGTAATGACCGTCTTAACATTAGCTCTCAACGCTATGCAGTGGCTTTGGTGTCGCTTACTACCAGTTCCGAACGTCTCGGGGGTCTTGATGCAGCGCACCACGATGCCAAGCCACGCTGAATCGGGCAGCCAGCAGGATCGTTACAGTCGAGCCCACGGCAATGCGAGCCGTCACGTTGCCGAATATGCACTGCGGGACTACGGACGCAGGGCCGCCAACCATCTGGTTGACGATAGCCAAACACGGTCCAGAGGTGCCCCATCTCGGTTATTTGGATCGTAATTTTTGTGAAGAACACTGCAAACGCAAGTGGCTTTCCAGACTTTACGACTATGTGCCATCATTTAAGGCTACGACCCGGCTTTTAAGACCCCCACCACTAAATAGAGGTACATCTGA'
    z_long = '3333321021021021021021021021021021021021021021021021021021021021021021033333333334564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564563210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210321021021021021021021021033334564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564563333333456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456332102102102102102102102102102102102102102102102102102102102102102102102102102102102102102102102103210210210210210210210210210210210210210210210210210210210210210'
    x_short = 'GTTTCCCAGTGTATATCGAGGGATACTACGTGCATAGTAACATCGGCCAA'
    z_short = '33333333333321021021021021021021021021021021021021' 

    genome_train = read_fasta_file_as_string("genome1", 0)
    true_ann_train_meta = read_fasta_file_as_string("true-ann1", 0)
    true_ann_train = translate_meta_states_to_states(true_ann_train_meta)

    genome_val = read_fasta_file_as_string("genome2", 100000)
    true_ann_val_meta = read_fasta_file_as_string("true-ann2", 100000)
    true_ann_val = translate_meta_states_to_states(true_ann_val_meta)
    
    print("These are the lengths:", len(genome_train), len(genome_val))
    print("These are ann lengths:")
    print("train ann:", len(true_ann_train_meta), len(true_ann_train))
    print("val ann:", len(true_ann_val_meta), len(true_ann_val))
    
    #train on that data
    hmm.training_by_counting_full(genome_train, true_ann_train)
    
    valid_after_train = validate_hmm(hmm)
    if not valid_after_train:
        print("Model is not valid after training - something is wrong!!!")
        print("trans_probs")
        print(hmm.trans_probs)
        print("emission_probs")
        print(hmm.emission_probs)
        quit()
    print("trans_probs")
    print(hmm.trans_probs)
    print("emission_probs")
    print(hmm.emission_probs)
    #do some decoding
    now = datetime.now()
    current_time = now.strftime("%H:%M:%S")
    print("Current Time =", current_time)
    w = hmm.compute_w_log_opt(genome_val)
    
    now = datetime.now()
    current_time = now.strftime("%H:%M:%S")
    print("Current Time =", current_time)

    w2 = hmm.compute_w_log(genome_val)

    now = datetime.now()

    current_time = now.strftime("%H:%M:%S")
    print("Current Time =", current_time)

    z_viterbi_log = hmm.backtrack_log_opt(genome_val, w)
    z_viterbi_log2 = hmm.backtrack_log(genome_val,w2)
    ww = hmm.opt_path_prob_log(w)
    ww2 = hmm.opt_path_prob_log(w2)

    print("viterbi_log:", ww, ww2)

    #acc = compute_accuracy(true_ann_val, z_viterbi_log)
    #print("Accuracy:", acc)



    '''
    #--------------------------------------------------------------------------------------#
    hmm = makeHmm()
    valid = validate_hmm(hmm)
    if not valid:
        print("Model is not valid - something is wrong!!!")
        quit()

    #train on 4 and validate on 5th
    size = 0

    #gotta count 'em aaa-aalll!
    A_max = np.zeros([hmm.K, hmm.K])
    fi_max = np.zeros([hmm.K, hmm.D])
    best_acc = 0
    for i in range(1, 6): #The genome number we validate on
        A_count = np.zeros([hmm.K, hmm.K])
        fi_count = np.zeros([hmm.K, hmm.D])
        for j in range(1, 6):  #The genome numbers we train on
            if i == j:
                continue
            genome_train1 = read_fasta_file_as_string("genome" + str(j), size)
            true_ann_train_meta1 = read_fasta_file_as_string("true-ann" + str(j), size)
            true_ann_train1 = translate_meta_states_to_states(true_ann_train_meta1)
            A_count1, fi_count1 = hmm.count_transitions_and_emissions(genome_train1, true_ann_train1)
            A_count += A_count1
            fi_count += fi_count1
            print("Finished count", j)
        
        #finish training
        hmm.training_by_counting(A_count, fi_count)

        valid_after_train = validate_hmm(hmm)
        if not valid_after_train:
            print("Model is not valid after training - something is wrong!!!")
            print("trans_probs")
            print(hmm.trans_probs)
            print("emission_probs")
            print(hmm.emission_probs)
            quit()

        #now validate
        genome_val = read_fasta_file_as_string("genome" + str(i), size)
        true_ann_val_meta = read_fasta_file_as_string("true-ann" + str(i), size)
        true_ann_val = translate_meta_states_to_states(true_ann_val_meta)
        
        w = hmm.compute_w_log(genome_val)
        z_viterbi_log = hmm.backtrack_log(genome_val, w)
        
        acc = compute_accuracy(true_ann_val, z_viterbi_log)
        print("Accuracy:", acc)

        if acc > best_acc:
            best_acc = acc 
            A_max = hmm.trans_probs
            fi_max = hmm.emission_probs
        
    print("Best accuracy: ", best_acc)

    hmm.trans_probs = A_max 
    hmm.emission_probs = fi_max 

    #decode
    for i in range(6,11):
        genome_decode = read_fasta_file_as_string("genome" + str(i), size)
        w = hmm.compute_w_log(genome_decode)
        z_viterbi_log = hmm.backtrack_log(genome_decode, w)
        z_meta = translate_states_to_meta_states(z_viterbi_log)
        #somehow make it into a fasta file
        path = "ASSIGNMENTS2021\Machine Learning\Assignment 3\FastaFiles\Results"
        filename = "fasta" + str(i) + ".fa"
        fullname = os.path.join(path,filename)
        file = open(fullname, "w")
        file.write(z_meta)
    '''
    
    

