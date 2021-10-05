import numpy as np
from h1_util import numerical_grad_check

def logistic(z):
    """ 
    Helper function
    Computes the logistic function 1/(1+e^{-x}) to each entry in input vector z.
    
    np.exp may come in handy
    Args:
        z: numpy array shape (d,) 
    Returns:
       logi: numpy array shape (d,) each entry transformed by the logistic function 
    """
    logi = np.zeros(z.shape)
    ### YOUR CODE HERE 1-5 lines
    logi = 1 / (1 + np.exp(- z))
    ### END CODE
    assert logi.shape == z.shape
    return logi

def permute(X, y):
    assert y.shape[0] == X.shape[0]
    xy = np.hstack((X,y[:, np.newaxis]))
    perm = np.random.permutation(xy)
    perm_y = perm[:,-1]
    perm_x = perm[:,:-1]
    assert y.shape == perm_y.shape
    assert X.shape == perm_x.shape
    return perm_x, perm_y


class LogisticRegressionClassifier():

    def __init__(self):
        self.w = None

    def cost_grad(self, X, y, w):
        """
        Compute the average negative log likelihood and gradient under the logistic regression model 
        using data X, targets y, weight vector w 
        
        np.log, np.sum, np.choose, np.dot may be useful here
        Args:
           X: np.array shape (n,d) float - Features 
           y: np.array shape (n,)  int - Labels 
           w: np.array shape (d,)  float - Initial parameter vector

        Returns:
           cost: scalar: the average negative log likelihood for logistic regression with data X, y 
           grad: np.arrray shape(d, ) gradient of the average negative log likelihood at w 
        """
        
        cost = 0
        grad = np.zeros(w.shape)
        ### YOUR CODE HERE 5 - 15 lines
        n = X.shape[0]

        cost_v = np.zeros(n)
        cost_v = -(y[:, np.newaxis] * (X @ w[:, np.newaxis]))
        cost_v = np.log(1 + np.exp(cost_v))
        cost = sum(cost_v) / n

        grad = logistic(-(y[:, np.newaxis] * (X @ w[:, np.newaxis])))
        grad_m = (-y[:, np.newaxis]) * X * grad #grad_m is nxd matrix
        grad = (np.sum(grad_m, axis=0) / n)
        ### END CODE
        assert grad.shape == w.shape
        return cost, grad


    def fit(self, X, y, w=None, lr=0.1, batch_size=16, epochs=10):
        """
        Run mini-batch stochastic Gradient Descent for logistic regression 
        use batch_size data points to compute gradient in each step.
    
        The function np.random.permutation may prove useful for shuffling the data before each epoch
        It is wise to print the performance of your algorithm at least after every epoch to see if progress is being made.
        Remeber the stochastic nature of the algorithm may give fluctuations in the cost as iterations increase.

        Args:
           X: np.array shape (n,d) dtype float32 - Features
           y: np.array shape (n,) dtype int32 - Labels
           w: np.array shape (d,) dtype float32 - Initial parameter vector
           lr: scalar - learning rate for gradient descent
           batch_size: number of elements to use in minibatch
           epochs: Number of scans through the data

        sets: 
           w: numpy array shape (d,) learned weight vector w
           history: list/np.array len epochs - value of cost function after every epoch. You know for plotting
        """
        if w is None: w = np.zeros(X.shape[1])
        history = []        
        ### YOUR CODE HERE 14 - 20 lines
        
        for e in range(epochs):
            perm_x, perm_y = permute(X, y)
            #print("shape permx, permy:", perm_x.shape, perm_y.shape)
            batchesX = [perm_x[i:i+batch_size] for i in range(0, len(perm_x), batch_size)]
            batchesY = [perm_y[i:i+batch_size] for i in range(0, len(perm_y), batch_size)]
            #OMGListsAreTheBest #Numpy4Life #I<3Python #LinearAlgebraAmirite 
            #oh, nothing on that index? sure, all good, have a nice day!
            latest_cost = 0
            for j in range(len(batchesX)):
                currentX = batchesX[j]
                currentY = batchesY[j]
                cost, grad = self.cost_grad(currentX, currentY, w)
                latest_cost = cost
                w = w - (lr * grad)
            print("Cost in epoch", e, ":", latest_cost)
            history.append(latest_cost)
        ### END CODE
        self.w = w
        self.history = history

    def predict(self, X):
        """ Classify each data element in X

        Args:
            X: np.array shape (n,d) dtype float - Features 
        
        Returns: 
           p: numpy array shape (n, ) dtype int32, class predictions on X (-1, 1)

        """
        out = np.ones(X.shape[0])
        ### YOUR CODE HERE 1 - 4 lines
        out = logistic(X @ self.w.T)   
        out = [(-1 if i < 0.5 else 1) for i in out]
        ### END CODE
        return out
    
    def score(self, X, y):
        """ Compute model accuracy  on Data X with labels y

        Args:
            X: np.array shape (n,d) dtype float - Features 
            y: np.array shape (n,) dtype int - Labels 

        Returns: 
           s: float, number of correct prediction divided by n.

        """
        s = 0
        ### YOUR CODE HERE 1 - 4 lines
        prediction = self.predict(X)
        s = np.sum(prediction == y) / len(y)
        ### END CODE
        return s
        

    
def test_logistic():
    print('*'*5, 'Testing logistic function')
    a = np.array([0, 1, 2, 3])
    lg = logistic(a)
    target = np.array([ 0.5, 0.73105858, 0.88079708, 0.95257413])
    assert np.allclose(lg, target), 'Logistic Mismatch Expected {0} - Got {1}'.format(target, lg)
    print('Logistic Test Success')

    
def test_cost():
    print('*'*5, 'Testing Cost Function')
    X = np.array([[1.0, 0.0], [1.0, 1.0], [3, 2]])
    y = np.array([-1, -1, 1], dtype='int64')
    w = np.array([0.0, 0.0])
    print('shapes', X.shape, w.shape, y.shape)
    lr = LogisticRegressionClassifier()
    cost,_ = lr.cost_grad(X, y, w)
    target = -np.log(0.5)
    assert np.allclose(cost, target), 'Cost Function Error:  Expected {0} - Got {1}'.format(target, cost)
    print('Cost Test Success')

    
def test_grad():
    print('*'*5, 'Testing  Gradient')
    X = np.array([[1.0, 0.0], [1.0, 1.0], [2.0, 3.0]])    
    w = np.array([0.0, 0.0])
    y = np.array([-1, -1, 1]).astype('int64')
    print('shapes', X.shape, w.shape, y.shape)
    lr = LogisticRegressionClassifier()
    f = lambda z: lr.cost_grad(X, y, w=z)
    numerical_grad_check(f, w)
    print('Grad Test Success')

def test_predict():
    lr = LogisticRegressionClassifier()
    lr.w = np.array([1.0,1.0,1.0])
    X = np.array([[1.0,2.0,-3.0],[-4.0,-5.0,-6.0]])
    out = lr.predict(X)
    check = out == np.array([1,-1])
    assert check.all()
    print('Predict Test Success')

def test_score():
    lr = LogisticRegressionClassifier()
    X = np.array([[1.0,2.0,-3.0],[-4.0,-5.0,-6.0]])
    lr.w = np.array([1.0,1.0,1.0])
    y = np.array([1.0,1.0])
    s = lr.score(X, y)
    assert 0.5 == s
    print('Score Test Success')
    
if __name__ == '__main__':
    test_logistic()
    test_predict()
    test_score() 
    test_cost()
    test_grad()
   
    

    
    
    
