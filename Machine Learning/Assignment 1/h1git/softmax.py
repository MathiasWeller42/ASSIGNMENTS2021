import numpy as np
from h1_util import numerical_grad_check

def softmax(X):
    """ 
    Compute the softmax of each row of an input matrix (2D numpy array).
    
    the numpy functions amax, log, exp, sum may come in handy as well as the keepdims=True option and the axis option.
    Remember to handle the numerical problems as discussed in the description.
    You should compute lg softmax first and then exponentiate 
    
    More precisely this is what you must do.
    
    For each row x do:
    compute max of x
    compute the log of the denominator sum for softmax but subtracting out the max i.e (log sum exp x-max) + max
    compute log of the softmax: x - logsum
    exponentiate that
    
    You can do all of it without for loops using numpys vectorized operations.

    Args:
        X: numpy array shape (n, d) each row is a data point
    Returns:
        res: numpy array shape (n, d)  where each row is the softmax transformation of the corresponding row in X i.e res[i, :] = softmax(X[i, :])
    """
    res = np.zeros(X.shape)
    ### YOUR CODE HERE no for loops please
    maxes = (np.amax(X, axis=1))[:, np.newaxis]
    lessMax_X = X - maxes
    exp_X = np.exp(lessMax_X)
    sum_X = np.sum(exp_X, axis=1)[:, np.newaxis] #this is a column vector
    log_X = np.log(sum_X) + maxes
    res = X - log_X
    res = np.exp(res)
    ### END CODE
    return res

def permute(X, y):
    assert y.shape[0] == X.shape[0]
    xy = np.hstack((X,y[:, np.newaxis]))
    perm = np.random.permutation(xy)
    perm_y = perm[:,-1]
    perm_x = perm[:,:-1]
    assert y.shape == perm_y.shape
    assert X.shape == perm_x.shape
    return perm_x, perm_y

def one_in_k_encoding(vec, k):
    """ One-in-k encoding of vector to k classes 
    
    Args:
       vec: numpy array - data to encode
       k: int - number of classes to encode to (0,...,k-1)
    """
    n = vec.shape[0]
    enc = np.zeros((n, k))
    enc[np.arange(n), vec] = 1
    return enc
    
class SoftmaxClassifier():

    def __init__(self, num_classes):
        self.num_classes = num_classes
        self.W = None
        
    def cost_grad(self, X, y, W):
        """ 
        Compute the average negative log likelihood cost and the gradient under the softmax model 
        using data X, Y and weight vector W.
        
        the functions np.log, np.nonzero, np.sum, np.dot (@), may come in handy
        Args:
           X: numpy array shape (n, d) float - the data each row is a data point
           y: numpy array shape (n, ) int - target values in 0,1,...,k-1
           W: numpy array shape (d x K) float - weight matrix
        Returns:
            totalcost: Average Negative Log Likelihood of w 
            gradient: The gradient of the average Negative Log Likelihood at w 
        """
        cost = np.nan
        grad = np.zeros(W.shape)*np.nan
        Yk = one_in_k_encoding(y, self.num_classes) # may help - otherwise you may remove it (nxK matrix)
        ### YOUR CODE HERE
        n = X.shape[0]
        soft_res = softmax(X @ W) #nxK matrix
        only_ys = [soft_res[i, y[i]] for i in range(n)]
        log_res = np.log(only_ys)
        cost = - 1.0/float(n) * np.sum(log_res)

        #Luckily, we are told that the gradient is:
        grad = -1.0/float(n) * X.T @ (Yk - softmax(X@W))
        assert grad.shape == W.shape
        ### END CODE
        return cost, grad


    def fit(self, X, Y, W=None, lr=0.01, epochs=10, batch_size=16):
        """
        Run Mini-Batch Gradient Descent on data X,Y to minimize the in sample error (1/n)NLL for softmax regression.
        Printing the performance every epoch is a good idea to see if the algorithm is working
    
        Args:
           X: numpy array shape (n, d) - the data each row is a data point
           Y: numpy array shape (n,) int - target labels numbers in {0, 1,..., k-1}
           W: numpy array shape (d x K)
           lr: scalar - initial learning rate
           batchsize: scalar - size of mini-batch
           epochs: scalar - number of iterations through the data to use

        Sets: 
           W: numpy array shape (d, K) learned weight vector matrix  W
           history: list/np.array len epochs - value of cost function after every epoch. You know for plotting
        """
        if W is None: W = np.zeros((X.shape[1], self.num_classes))
        history = []
        ### YOUR CODE HERE
        for e in range(epochs):
            perm_x, perm_y = permute(X, Y)
            #print("shape permx, permy:", perm_x.shape, perm_y.shape)
            batchesX = [perm_x[i:i+batch_size] for i in range(0, len(perm_x), batch_size)]
            batchesY = [perm_y[i:i+batch_size] for i in range(0, len(perm_y), batch_size)]
            latest_cost = 0
            for j in range(len(batchesX)):
                currentX = batchesX[j]
                currentY = batchesY[j].astype('int64')
                cost, grad = self.cost_grad(currentX, currentY, W)
                latest_cost = cost
                W = W - (lr * grad)
            print("Cost in epoch", e, ":", latest_cost)
            history.append(latest_cost)
        ### END CODE
        self.W = W
        self.history = history
        

    def score(self, X, Y):
        """ Compute accuracy of classifier on data X with labels Y

        Args:
           X: numpy array shape (n, d) - the data each row is a data point
           Y: numpy array shape (n,) int - target labels numbers in {0, 1,..., k-1}
        Returns:
           out: float - mean accuracy
        """
        out = 0
        ### YOUR CODE HERE 1-4 lines
        prediction = self.predict(X)
        out = np.sum(prediction == Y) / len(Y)
        ### END CODE
        return out

    def predict(self, X):
        """ Compute classifier prediction on each data point in X 

        Args:
           X: numpy array shape (n, d) - the data each row is a data point
        Returns
           out: np.array shape (n, ) - prediction on each data point (number in 0,1,..., num_classes)
        """
        out = None
        ### YOUR CODE HERE - 1-4 lines
        out = np.zeros(X.shape[0])
        softmax_res = softmax(X @ self.W) 
        out = np.argmax(softmax_res, axis=1)
        assert out.shape[0] == X.shape[0]
        ### END CODE
        return out


def test_encoding():
    print('*'*10, 'test encoding')
    labels = np.array([0, 2, 1, 1])
    m = one_in_k_encoding(labels, 3)
    res =  np.array([[1, 0 , 0], [0, 0, 1], [0, 1, 0], [0, 1, 0]])
    assert res.shape == m.shape, 'encoding shape mismatch'
    assert np.allclose(m, res), m - res
    print('Test Passed')

    
def test_softmax():
    print('Test softmax')
    X = np.zeros((3 ,2))
    X[0, 0] = np.log(4)
    X[1, 1] = np.log(2)
    print('Input to Softmax: \n', X)
    sm = softmax(X)
    expected = np.array([[4.0/5.0, 1.0/5.0], [1.0/3.0, 2.0/3.0], [0.5, 0.5]])
    print('Result of softmax: \n', sm)
    assert np.allclose(expected, sm), 'Expected {0} - got {1}'.format(expected, sm)
    print('Test complete')
        

def test_grad():
    scl = SoftmaxClassifier(num_classes=3)
    X = np.array([[1.0, 0.0], [1.0, 1.0], [1.0, -1.0]])    
    w = np.ones((2, 3))
    y = np.array([0, 1, 2])
    print('*'*5, 'Testing  Gradient')
    f = lambda z: scl.cost_grad(X, y, W=z)
    numerical_grad_check(f, w)
    print('Grad test Success')

def test_cost():
    print('*'*5, 'Testing  Cost')
    scl = SoftmaxClassifier(num_classes=3)
    X = np.array([[1.0, 0.0], [1.0, 1.0], [1.0, -1.0]])    
    w = np.ones((2, 3))
    y = np.array([0, 1, 2])
    cost, _ = scl.cost_grad(X, y, w)
    assert np.allclose(cost,np.log(3)), 'Expected {0} - got {1}'.format(np.log(3), cost)
    print('Cost test success')


def test_predict():
    cl = SoftmaxClassifier(num_classes=2)
    x = np.array([[7,2],[6,3],[5,6]])
    cl.W = np.array([[2,1],[3,4]])
    out = cl.predict(x)
    assert np.allclose(out, np.array([0,0,1]))
    print("Prediction test passed")

def test_score():
    cl = SoftmaxClassifier(num_classes=2)
    x = np.array([[7,2],[6,3],[5,6]])
    cl.W = np.array([[2,1],[3,4]])
    y = np.array([0,0,1])
    res = cl.score(x,y)    
    assert res == 1
    
    y = np.array([0,1,1])
    res = cl.score(x,y)    
    assert res == (2/3)
    print("Score test passed")

if __name__ == "__main__":
    test_encoding()
    test_softmax()
    test_score()
    test_predict()
    test_cost()
    test_grad()



    
    

