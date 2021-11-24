import numpy as np
################################ FUNCTIONS WE ARE GIVEN#############################################
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


def translate_indices_to_observations(indices):
    mapping = ['a', 'c', 'g', 't']
    return ''.join(mapping[idx] for idx in indices)

def translate_path_to_indices(path):
    return list(map(lambda x: int(x), path))

def translate_indices_to_path(indices):
    return ''.join([str(i) for i in indices])

class hmm:
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
        
    def training_by_counting(self, x, z):
        """
        Returns a HMM trained on x and z cf. training-by-counting.
        """
        A, phi = self.count_transitions_and_emissions(x, z)
        for i in range(A.shape[0]):
            sum = np.sum(A[i,:])
            A[i,:] /= sum
            sum2 = np.sum(phi[i,:])
            phi[i,:] /= sum2
        self.emission_probs = phi
        self.trans_probs = A
        

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
    pi =np.array( model.init_probs)
    A = np.array(model.trans_probs)
    fi =np.array(model.emission_probs)
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


if __name__ == '__main__':
    #make (and validate) the model
    hmm = makeHmm()
    valid = validate_hmm(hmm)
    if not valid:
        print("Model is not valid - something is wrong!!!")

    #read some data
    x_long = 'TGAGTATCACTTAGGTCTATGTCTAGTCGTCTTTCGTAATGTTTGGTCTTGTCACCAGTTATCCTATGGCGCTCCGAGTCTGGTTCTCGAAATAAGCATCCCCGCCCAAGTCATGCACCCGTTTGTGTTCTTCGCCGACTTGAGCGACTTAATGAGGATGCCACTCGTCACCATCTTGAACATGCCACCAACGAGGTTGCCGCCGTCCATTATAACTACAACCTAGACAATTTTCGCTTTAGGTCCATTCACTAGGCCGAAATCCGCTGGAGTAAGCACAAAGCTCGTATAGGCAAAACCGACTCCATGAGTCTGCCTCCCGACCATTCCCATCAAAATACGCTATCAATACTAAAAAAATGACGGTTCAGCCTCACCCGGATGCTCGAGACAGCACACGGACATGATAGCGAACGTGACCAGTGTAGTGGCCCAGGGGAACCGCCGCGCCATTTTGTTCATGGCCCCGCTGCCGAATATTTCGATCCCAGCTAGAGTAATGACCTGTAGCTTAAACCCACTTTTGGCCCAAACTAGAGCAACAATCGGAATGGCTGAAGTGAATGCCGGCATGCCCTCAGCTCTAAGCGCCTCGATCGCAGTAATGACCGTCTTAACATTAGCTCTCAACGCTATGCAGTGGCTTTGGTGTCGCTTACTACCAGTTCCGAACGTCTCGGGGGTCTTGATGCAGCGCACCACGATGCCAAGCCACGCTGAATCGGGCAGCCAGCAGGATCGTTACAGTCGAGCCCACGGCAATGCGAGCCGTCACGTTGCCGAATATGCACTGCGGGACTACGGACGCAGGGCCGCCAACCATCTGGTTGACGATAGCCAAACACGGTCCAGAGGTGCCCCATCTCGGTTATTTGGATCGTAATTTTTGTGAAGAACACTGCAAACGCAAGTGGCTTTCCAGACTTTACGACTATGTGCCATCATTTAAGGCTACGACCCGGCTTTTAAGACCCCCACCACTAAATAGAGGTACATCTGA'
    z_long = '3333321021021021021021021021021021021021021021021021021021021021021021033333333334564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564563210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210210321021021021021021021021033334564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564564563333333456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456456332102102102102102102102102102102102102102102102102102102102102102102102102102102102102102102102103210210210210210210210210210210210210210210210210210210210210210'

    #train on that data
    print(hmm.count_transitions_and_emissions(x_long,z_long))
    hmm.training_by_counting(x_long, z_long)
    print(hmm.trans_probs)
    print(hmm.emission_probs)
    
    

