import numpy as np
from datetime import datetime
import scipy.optimize
import copy
from fractions import Fraction
from enum import Enum

def example1(): return np.array([5,4,3]),np.array([[2,3,1],[4,1,2],[3,4,2]]),np.array([5,11,8])
def example2(): return np.array([-2,-1]),np.array([[-1,1],[-1,-2],[0,1]]),np.array([-1,-2,1])
def integer_pivoting_example(): return np.array([5,2]),np.array([[3,1],[2,5]]),np.array([7,5])
def exercise2_5(): return np.array([1,3]),np.array([[-1,-1],[-1,1],[1,2]]),np.array([-3,-1,4])
def exercise2_6(): return np.array([1,3]),np.array([[-1,-1],[-1,1],[1,2]]),np.array([-3,-1,2])
def exercise2_7(): return np.array([1,3]),np.array([[-1,-1],[-1,1],[-1,2]]),np.array([-3,-1,2])
def random_lp(n,m,sigma=10): return np.round(sigma*np.random.randn(n)),np.round(sigma*np.random.randn(m,n)),np.round(sigma*np.abs(np.random.randn(m)))

def exercise2_1(): return np.array([6,8,5,9]),np.array([[2,1,1,3],[1,3,1,2]]),np.array([5,3])
def exercise2_7_after_aux(): return np.array([1,2]),np.array([[-1,-1],[-1,2],[-1,3]]),np.array([3,2,5])

class Dictionary:
    # Simplex dictionary as defined by Vanderbei
    # 'C' is a (m+1)x(n+1) NumPy array that stores all the coefficients
    # of the dictionary.
    #
    # 'dtype' is the type of the entries of the dictionary. It is
    # supposed to be one of the native (full precision) Python types
    # 'int' or 'Fraction' or any Numpy type such as 'np.float64'.
    #
    # dtype 'int' is used for integer pivoting. Here an additional
    # variables 'lastpivot' is used. 'lastpivot' is the negative pivot
    # coefficient of the previous pivot operation. Dividing all
    # entries of the integer dictionary by 'lastpivot' results in the
    # normal dictionary.
    #
    # Variables are indexed from 0 to n+m. Variable 0 is the objective
    # z. Variables 1 to n are the original variables. Variables n+1 to
    # n+m are the slack variables. An exception is when creating an
    # auxillary dictionary where variable n+1 is the auxillary
    # variable (named x0) and variables n+2 to n+m+1 are the slack
    # variables (still names x{n+1} to x{n+m}).
    #
    # 'B' and 'N' are arrays that contain the *indices* of the basic and
    # nonbasic variables.
    #
    # 'varnames' is an array of the names of the variables.
    
    def __init__(self,c,A,b,dtype=Fraction):
        # Initializes the dictionary based on linear program in
        # standard form given by vectors and matrices 'c','A','b'.
        # Dimensions are inferred from 'A' 
        #
        # If 'c' is None it generates the auxillary dictionary for the
        # use in the standard two-phase simplex algorithm
        #
        # Every entry of the input is individually converted to the
        # given dtype.
        m,n = A.shape
        self.dtype=dtype
        if dtype == int:
            self.lastpivot=1
        if dtype in [int,Fraction]:
            dtype=object
            if c is not None:
                c=np.array(c,np.object)
            A=np.array(A,np.object)
            b=np.array(b,np.object)
        self.C = np.empty([m+1,n+1+(c is None)],dtype=dtype)
        self.C[0,0]=self.dtype(0)
        if c is None:
            self.C[0,1:]=self.dtype(0)
            self.C[0,n+1]=self.dtype(-1)
            self.C[1:,n+1]=self.dtype(1)
        else:
            for j in range(0,n):
                self.C[0,j+1]=self.dtype(c[j])
        for i in range(0,m):
            self.C[i+1,0]=self.dtype(b[i])
            for j in range(0,n):
                self.C[i+1,j+1]=self.dtype(-A[i,j])
        self.N = np.array(range(1,n+1+(c is None)))
        self.B = np.array(range(n+1+(c is None),n+1+(c is None)+m))
        self.varnames=np.empty(n+1+(c is None)+m,dtype=object)
        self.varnames[0]='z'
        for i in range(1,n+1):
            self.varnames[i]='x{}'.format(i)
        if c is None:
            self.varnames[n+1]='x0'
        for i in range(n+1,n+m+1):
            self.varnames[i+(c is None)]='x{}'.format(i)

    def __str__(self):
        # String representation of the dictionary in equation form as
        # used in Vanderbei.
        m,n = self.C.shape
        varlen = len(max(self.varnames,key=len))
        coeflen = 0
        for i in range(0,m):
            coeflen=max(coeflen,len(str(self.C[i,0])))
            for j in range(1,n):
                coeflen=max(coeflen,len(str(abs(self.C[i,j]))))
        tmp=[]
        if self.dtype==int and self.lastpivot!=1:
            tmp.append(str(self.lastpivot))
            tmp.append('*')
        tmp.append('{} = '.format(self.varnames[0]).rjust(varlen+3))
        tmp.append(str(self.C[0,0]).rjust(coeflen))
        for j in range(0,n-1):
            tmp.append(' + ' if self.C[0,j+1]>0 else ' - ')
            tmp.append(str(abs(self.C[0,j+1])).rjust(coeflen))
            tmp.append('*')
            tmp.append('{}'.format(self.varnames[self.N[j]]).rjust(varlen))
        for i in range(0,m-1):
            tmp.append('\n')
            if self.dtype==int and self.lastpivot!=1:
                tmp.append(str(self.lastpivot))
                tmp.append('*')
            tmp.append('{} = '.format(self.varnames[self.B[i]]).rjust(varlen+3))
            tmp.append(str(self.C[i+1,0]).rjust(coeflen))
            for j in range(0,n-1):
                tmp.append(' + ' if self.C[i+1,j+1]>0 else ' - ')
                tmp.append(str(abs(self.C[i+1,j+1])).rjust(coeflen))
                tmp.append('*')
                tmp.append('{}'.format(self.varnames[self.N[j]]).rjust(varlen))
        return ''.join(tmp)

    def basic_solution(self):
        # Extracts the basic solution defined by a dictionary D
        m,n = self.C.shape
        if self.dtype==int:
            x_dtype=Fraction
        else:
            x_dtype=self.dtype
        x = np.empty(n-1,x_dtype)
        x[:] = x_dtype(0)
        for i in range (0,m-1):
            if self.B[i]<n:
                if self.dtype==int:
                    x[self.B[i]-1]=Fraction(self.C[i+1,0],self.lastpivot)
                else:
                    x[self.B[i]-1]=self.C[i+1,0]
        return x

    def value(self):
        # Extracts the value of the basic solution defined by a dictionary D
        if self.dtype==int:
            return Fraction(self.C[0,0],self.lastpivot)
        else:
            return self.C[0,0]

    def pivot(self,k,l):
        # Pivot Dictionary with N[k] entering and B[l] leaving
        # Performs integer pivoting if self.dtype==int

        # save pivot coefficient column of the pivot
        pivot_coeff = self.C[l+1,k+1]
        nonbasic_col = self.C[:, k+1].copy()
        
        # set the column of the pivot to 0 (needed for row updates later)
        self.C[:, k+1] = 0

        # update the row of the pivot
        self.C[l+1,k+1] = -1
        self.C[l+1,:] = self.C[l+1,:] / (-pivot_coeff)

        # update all other rows of C
        for i in range(0, self.C.shape[0]):
            if i == l+1:
                continue
            coeff = nonbasic_col[i]
            self.C[i,:] = self.C[i,:] + coeff * self.C[l+1,:]
        
        # update N and B
        nonbasic = self.N[k]
        self.N[k] = self.B[l]
        self.B[l] = nonbasic

class LPResult(Enum):
    OPTIMAL = 1
    INFEASIBLE = 2
    UNBOUNDED = 3

def bland(D,eps):
    # Assumes a feasible dictionary D and finds entering and leaving
    # variables according to Bland's rule.
    #
    # eps>=0 is such that numbers in the closed interval [-eps,eps]
    # are to be treated as if they were 0
    #
    # Returns k and l such that
    # k is None if D is Optimal
    # Otherwise D.N[k] is entering variable
    # l is None if D is Unbounded
    # Otherwise D.B[l] is a leaving variable
       
    k=l=None
    # find possible entering variables
    nonbasics = D.C[0,1:]    
    possible_entering_indexes = [i for i, e in enumerate(nonbasics) if e > eps]

    # if the list is empty (optimal) return immediately
    if len(possible_entering_indexes) == 0:
        return k,l

    # otherwise find the best entering according to D.N
    variable_indexes = D.N[possible_entering_indexes]
    k = possible_entering_indexes[np.argmin(variable_indexes)]
    # then find possible leaving variables
    nonbasic_col = D.C[1:, k+1]
    b_values = D.C[1:, 0]
    fractions = [(i, handleFraction(a,b,eps)) for (i, (b,a)) in enumerate(zip(b_values, nonbasic_col))]

    #Handle unboundedness:
    if np.all([a >= -eps for a in nonbasic_col]):
        return k, None

    smallest_fraction = np.amin([f for (_,f) in fractions if f>= -eps])

    indexes_of_smallest_fraction = [i for (i,f) in fractions if f == smallest_fraction]

    variable_indexes_with_smallest_fraction = D.B[indexes_of_smallest_fraction]
    l = indexes_of_smallest_fraction[np.argmin(variable_indexes_with_smallest_fraction)]

    return k,l

def handleFraction(a,b, eps):
    #(i, b / (-a)) if a != 0 else (i,handleDivisionByZero(a,b))
    if a >= -eps and a <= eps:
        if b >= -eps and b <= eps:
            return 0
        else:
            return np.iinfo(int).max
    elif a < -eps:
        return b / (-a)
    else:
        if b >= -eps and b <= eps:
            return np.iinfo(int).min 
        else:
            return b / (-a)

def largest_coefficient(D,eps):
    # Assumes a feasible dictionary D and find entering and leaving
    # variables according to the Largest Coefficient rule.
    #
    # eps>=0 is such that numbers in the closed interval [-eps,eps]
    # are to be treated as if they were 0
    #
    # Returns k and l such that
    # k is None if D is Optimal
    # Otherwise D.N[k] is entering variable
    # l is None if D is Unbounded
    # Otherwise D.B[l] is a leaving variable
    
    k=l=None
    nonbasics = D.C[0,1:]    
    possible_entering = [(i,e) for i,e in enumerate(nonbasics) if e > eps]
    
    # if the list is empty (optimal) return immediately
    if len(possible_entering) == 0:
        return k,l

    # otherwise we now find the best entering according to D.N:
    k = max(possible_entering, key= lambda item: (item[1], -D.N[item[0]]))[0]
    
    # then find possible leaving variables (same as Bland)
    nonbasic_col = D.C[1:, k+1]
    b_values = D.C[1:, 0]
    fractions = [(i, handleFraction(a,b,eps)) for (i, (b,a)) in enumerate(zip(b_values, nonbasic_col))]

    #Handle unboundedness:
    if np.all([a >= -eps for a in nonbasic_col]):
        return k, None

    smallest_fraction = np.amin([f for (_,f) in fractions if f>= -eps])

    indexes_of_smallest_fraction = [i for (i,f) in fractions if f == smallest_fraction]

    variable_indexes_with_smallest_fraction = D.B[indexes_of_smallest_fraction]
    l = indexes_of_smallest_fraction[np.argmin(variable_indexes_with_smallest_fraction)]

    return k,l

def largest_increase(D,eps):
    # Assumes a feasible dictionary D and find entering and leaving
    # variables according to the Largest Increase rule.
    #
    # eps>=0 is such that numbers in the closed interval [-eps,eps]
    # are to be treated as if they were 0
    #
    # Returns k and l such that
    # k is None if D is Optimal
    # Otherwise D.N[k] is entering variable
    # l is None if D is Unbounded
    # Otherwise D.B[l] is a leaving variable
    
    k=l=None
    # TODO

    nonbasics = D.C[0,1:]    
    possible_entering = [i for i,e in enumerate(nonbasics) if e > eps]
    
    # if the list is empty (optimal) return immediately
    if len(possible_entering) == 0:
        return k,l

    b_values = D.C[1:, 0]
    fractions = np.zeros((possible_entering.shape[0], D.C.shape[1]-1))
    for (j, index) in enumerate(possible_entering):
        nonbasic_col = D.C[1:, index+1]
        fractions[:, j] = [handleFraction(a,b,eps) for (b,a) in zip(b_values, nonbasic_col)]
    
    #TODO check unboundedness

    fractions = fractions * D.C[0,1:][possible_entering]
    
    k,l = np.argmax(fractions)
    
    
    # then find possible leaving variables (same as Bland)
    #nonbasic_col = D.C[1:, k+1]
    #b_values = D.C[1:, 0]
    #fractions = [(i, handleFraction(a,b,eps)) for (i, (b,a)) in enumerate(zip(b_values, nonbasic_col))]

    #Handle unboundedness:
    if np.all([a >= -eps for a in nonbasic_col]):
        return k, None

    smallest_fraction = np.amin([f for (_,f) in fractions if f>= -eps])

    indexes_of_smallest_fraction = [i for (i,f) in fractions if f == smallest_fraction]

    variable_indexes_with_smallest_fraction = D.B[indexes_of_smallest_fraction]
    l = indexes_of_smallest_fraction[np.argmin(variable_indexes_with_smallest_fraction)]

    return k,l

def lp_solve(c,A,b,dtype=Fraction,eps=0.,pivotrule=lambda D,eps: bland(D,eps),verbose=False):
    # Simplex algorithm
    #    
    # Input is LP in standard form given by vectors and matrices
    # c,A,b.
    #
    # eps>=0 is such that numbers in the closed interval [-eps,eps]
    # are to be treated as if they were 0.
    #
    # pivotrule is a rule used for pivoting. Cycling is prevented by
    # switching to Bland's rule as needed.
    #
    # If verbose is True it outputs possible useful information about
    # the execution, e.g. the sequence of pivot operations
    # performed. Nothing is required.
    #
    # If LP is infeasible the return value is LPResult.INFEASIBLE,None
    #
    # If LP is unbounded the return value is LPResult.UNBOUNDED,None
    #
    # If LP has an optimal solution the return value is
    # LPResult.OPTIMAL,D, where D is an optimal dictionary.

    D=Dictionary(c,A,b,dtype)
    if verbose:
        print("Solving the dictionary:")
        print(D)
        print("Checking whether to run a dual problem...")
    
    bs_below_zero = b <= 0
    if bs_below_zero.any():
        if verbose:
            print("Initial problem infeasible. Running dual")
        #save the primal and create the dual
        oldDict = copy.deepcopy(D)
        D.C[0,:] = -np.ones(D.C.shape[1], dtype=dtype)
        D.C[0,0] = Fraction(0)
        D.C = -D.C.T
        oldN = D.N
        D.N = D.B
        D.B = oldN
        if verbose:
            print("The dual dictionary to solve:")
            print(D)
        result, dual = runSolve(D, eps, pivotrule, verbose)
        if result == LPResult.UNBOUNDED:
            if verbose:
                print("End of computation, result:")
            return LPResult.INFEASIBLE, None
        else:
            #setup to run the primal (phase 2)
            dual.C = -dual.C.T
            dualN = dual.N
            dual.N = dual.B
            dual.B = dualN
            dual.C[0,:] = np.zeros(dual.C.shape[1], dtype=dtype)
            for (i, coeff) in enumerate(oldDict.C[0, 1:]):
                currentOldVar = oldDict.N[i]
                if currentOldVar in dual.B:
                    index = np.where (dual.B == currentOldVar)[0][0]
                    dual.C[0,:] += coeff * dual.C[index + 1, :]
                else:
                    dual.C[0,i+1] += coeff
            if verbose:
                print("End of dual, now running phase 2")
                print("The primal dictionary looks like this:")
                print(D)
    elif verbose:
        print("Dual problem not needed")

    return runSolve(D,eps,pivotrule,verbose)        

def runSolve(D,eps=0.,pivotrule=lambda D,eps: bland(D,eps),verbose=False): 
    while True:
        k,l = pivotrule(D,eps)
        
        if k == None:
            if verbose:
                print("End of computation, result:")
            return LPResult.OPTIMAL, D 
        elif l == None:
            if verbose:
                print("End of computation, result:")
            return LPResult.UNBOUNDED, None 
        else:
            if verbose:
                print("Found pivots, entering: x", D.N[k], "leaving: x", D.B[l])
            D.pivot(k,l)
            if verbose:
                print("Dictionary after pivot:")
                print(D)
    
def run_examples():
    """
    # Example 1
    c,A,b = example1()
    D=Dictionary(c,A,b)
    print('Example 1 with Fraction')
    print('Initial dictionary:')
    print(D)
    print('x1 is entering and x4 leaving:')
    D.pivot(0,0)
    print(D)
    print('x3 is entering and x6 leaving:')
    D.pivot(2,2)
    print(D)
    print()

    D=Dictionary(c,A,b,np.float64)
    print('Example 1 with np.float64')
    print('Initial dictionary:')
    print(D)
    print('x1 is entering and x4 leaving:')
    D.pivot(0,0)
    print(D)
    print('x3 is entering and x6 leaving:')
    D.pivot(2,2)
    print(D)
    print()

    # Example 2
    c,A,b = example2()
    print('Example 2')
    print('Auxillary dictionary')
    D=Dictionary(None,A,b)
    print(D)
    print('x0 is entering and x4 leaving:')
    D.pivot(2,1)
    print(D)
    print('x2 is entering and x3 leaving:')
    D.pivot(1,0)
    print(D)
    print('x1 is entering and x0 leaving:')
    D.pivot(0,1)
    print(D)
    print()
    
"""
    # Solve Example 1 using lp_solve
    c,A,b = example1()
    D=Dictionary(c,A,b)
    print('lp_solve Example 1 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!:')
    print(D)
    res,D=lp_solve(c,A,b,pivotrule=lambda D,eps: largest_coefficient(D,eps),verbose=True)
    print(res)
    print(D)
    print()

    # Solve Example 2 using lp_solve 
    c,A,b = example2()
    print('lp_solve Example 2 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!:')
    res,D=lp_solve(c,A,b,pivotrule=lambda D,eps: largest_coefficient(D,eps),verbose=True)
    print(res)
    print(D)
    print()
    

    # Solve Exercise 2.5 using lp_solve 
    c,A,b = exercise2_5()
    print('lp_solve Exercise 2.5 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!:')
    res,D=lp_solve(c,A,b,verbose=False)
    print(res)
    print(D)
    print()


    # Solve Exercise 2.6 using lp_solve
    c,A,b = exercise2_6()
    print('lp_solve Exercise 2.6 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!:')
    res,D=lp_solve(c,A,b,verbose=False)
    print(res)
    print(D)
    print()

    # Solve Exercise 2.7 using lp_solve
    c,A,b = exercise2_7()
    print('lp_solve Exercise 2.7 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!:')
    res,D=lp_solve(c,A,b,verbose=False)
    print(res)
    print(D)
    print() 

'''
    #Integer pivoting
    c,A,b=example1()
    D=Dictionary(c,A,b,int)
    print('Example 1 with int')
    print('Initial dictionary:')
    print(D)
    print('x1 is entering and x4 leaving:')
    D.pivot(0,0)
    print(D)
    print('x3 is entering and x6 leaving:')
    D.pivot(2,2)
    print(D)
    print()

    c,A,b = integer_pivoting_example()
    D=Dictionary(c,A,b,int)
    print('Integer pivoting example from lecture')
    print('Initial dictionary:')
    print(D)
    print('x1 is entering and x3 leaving:')
    D.pivot(0,0)
    print(D)
    print('x2 is entering and x4 leaving:')
    D.pivot(1,1)
    print(D)
    '''


def run_examples_homebrew():
    #IPA
    '''c,A,b = example1()
    D = Dictionary(c,A,b)
    print(D)
    k,l = bland(D, 0)
    print("The variable number:", D.N[k])
    print("k: ", k, " and l: ", l)'''

    # Solve Example 2 using lp_solve 
    c,A,b = exercise2_1()
    print('lp_solve exercise 2.1 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!:')
    res,D=lp_solve(c,A,b)
    print(res)
    print(D)
    print()

    c,A,b = exercise2_7_after_aux()
    print('lp_solve exercise 2.7 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!:')
    res,D=lp_solve(c,A,b)
    print(res)
    print(D)
    print()

def experiment1():
    print("Testing implementation")
    time_frac = 0
    time_float = 0
    time_scipy = 0
    pivotrule = lambda D,eps : largest_coefficient(D,eps)
    verbose = False
    print("#################### Running FRACTIONS")
    for i in range(200): #fractions
        c,A,b = generateDictionary(i)
        starttime = datetime.now()
        res,_ = lp_solve(c,A,b,pivotrule=pivotrule,verbose=verbose)
        time_frac = time_frac + (datetime.now() - starttime).total_seconds()
        print("this is the FRACTION result: ", res, "and i:", i)
    
    print("#################### Running FLOATS")
    for i in range(200): #floats
        c,A,b = generateDictionary(i)
        starttime = datetime.now()
        D=Dictionary(c,A,b)
        #print("This is the new dictionary", i, ":", D)
        res,_ = lp_solve(c,A,b,dtype=np.float64,eps=0.001,pivotrule=pivotrule,verbose=verbose)
        time_float = time_float + (datetime.now() - starttime).total_seconds()
        print("this is the FLOAT result: ", res, "and i:", i)

    print("#################### Running SCIPY") 
    for i in range(200): #scipy
        c,A,b = generateDictionary(i)
        c = -c
        D=Dictionary(c,A,b)
        #print("This is the new dictionary", i, ":", D)
        starttime = datetime.now()
        res = scipy.optimize.linprog(c=c,A_ub=A,b_ub=b, method='simplex')
        time_scipy = time_scipy + (datetime.now() - starttime).total_seconds()
        print("this is the SCIPY res:", statusconv(res.status), " and i: ", i)
    
    print("Time for fraction implementation", time_frac)
    print("Time for float implementation:", time_float)
    print("Time for scipy implementation:", time_scipy)

def generateDictionary(seed):
    np.random.seed(seed)
    n = np.random.randint(2,3)
    m = np.random.randint(2,3)
    #print("This is n and m:", n, m)
    c = np.random.randint(0,20,size=n)
    b = np.random.randint(1,10,size=m)
    A = np.random.randint(-10,10,size=(m,n))
    return c,A,b

def statusconv(a):
    if a == 0:
        return LPResult.OPTIMAL
    elif a == 3:
        return LPResult.UNBOUNDED
    else:
        return LPResult.INFEASIBLE
    

if __name__ == "__main__":
    print("######################################### NEW RUN ##########################################################")
    #run_examples()
    #run_examples_homebrew()
    experiment1()



