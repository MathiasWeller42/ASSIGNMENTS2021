import numpy as np

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

def softmax(X):
    """ 
    You can take this from handin I
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
    ### YOUR CODE HERE
    maxes = (np.amax(X, axis=1))[:, np.newaxis]
    lessMax_X = X - maxes
    exp_X = np.exp(lessMax_X)
    sum_X = np.sum(exp_X, axis=1)[:, np.newaxis] 
    log_X = np.log(sum_X) + maxes
    res = X - log_X
    res = np.exp(res)
    ### END CODE
    return res

def relu(x):
    """ Compute the relu activation function on every element of the input
    
        Args:
            x: np.array
        Returns:
            res: np.array same shape as x
        Beware of np.max and look at np.maximum
    """
    ### YOUR CODE HERE
    res = np.maximum(x, np.zeros(x.shape))
    assert res.shape == x.shape
    ### END CODE
    return res

def make_dict(W1, b1, W2, b2):
    """ Trivial helper function """
    return {'W1': W1, 'b1': b1, 'W2': W2, 'b2': b2}


def get_init_params(input_dim, hidden_size, output_size):
    """ Initializer function using Xavier/he et al Delving Deep into Rectifiers: Surpassing Human-Level Performance on ImageNet Classification

    Args:
      input_dim: int
      hidden_size: int
      output_size: int
    Returns:
       dict of randomly initialized parameter matrices.
    """
    W1 = np.random.normal(0, np.sqrt(2./(input_dim+hidden_size)), size=(input_dim, hidden_size))
    b1 = np.zeros((1, hidden_size))
    W2 = np.random.normal(0, np.sqrt(4./(hidden_size+output_size)), size=(hidden_size, output_size))
    b2 = np.zeros((1, output_size))
    return {'W1': W1, 'b1': b1, 'W2': W2, 'b2': b2}

  
class NetClassifier():
    
    def __init__(self):
        """ Trivial Init """
        self.params = None
        self.hist = None

    def predict(self, X, params=None):
        """ Compute class prediction for all data points in class X
        
        Args:
            X: np.array shape n, d
            params: dict of params to use (if none use stored params)
        Returns:
            np.array shape n, 1
        """
        if params is None:
            params = self.params
        pred = None
        ### YOUR CODE HERE
        W1 = params['W1']
        b1 = params['b1']
        W2 = params['W2']
        b2 = params['b2']
        output1 = relu(X @ W1 + b1)
        output2 = output1 @ W2 + b2
        probabilities = softmax(output2) #Might have to remove this line?
        pred = np.argmax(probabilities, axis=1) 
        ### END CODE
        return pred
     
    def score(self, X, y, params=None):
        """ Compute accuracy of model on data X with labels y
        
        Args:
            X: np.array shape n, d
            y: np.array shape n, 1
            params: dict of params to use (if none use stored params)

        Returns:
            np.array shape n, 1
        """
        if params is None:
            params = self.params
        acc = None
        ### YOUR CODE HERE
        prediction = self.predict(X, params)
        acc = np.mean(prediction == y)
        ### END CODE
        return acc
    
    @staticmethod
    def cost_grad(X, y, params, c=0.0):
        """ Compute cost and gradient of neural net on data X with labels y using weight decay parameter c
        You should implement a forward pass and store the intermediate results 
        and the implement the backwards pass using the intermediate stored results
        
        Use the derivative for cost as a function for input to softmax as derived above
        
        Args:
            X: np.array shape n, self.input_size
            y: np.array shape n, 1
            params: dict with keys (W1, W2, b1, b2)
            c: float - weight decay parameter
            params: dict of params to use for the computation
        
        Returns 
            cost: scalar - average cross entropy cost
            dict with keys
            d_w1: np.array shape w1.shape, entry d_w1[i, j] = \partial cost/ \partial W1[i, j]
            d_w2: np.array shape w2.shape, entry d_w2[i, j] = \partial cost/ \partial W2[i, j]
            d_b1: np.array shape b1.shape, entry d_b1[1, j] = \partial cost/ \partial b1[1, j]
            d_b2: np.array shape b2.shape, entry d_b2[1, j] = \partial cost/ \partial b2[1, j]
            
        """
        
        W1 = params['W1']
        b1 = params['b1']
        W2 = params['W2']
        b2 = params['b2']
        labels = one_in_k_encoding(y, W2.shape[1]) # shape n x k
        #print("W1 dim:", W1.shape)
        #print("b1 dim:", b1.shape)
        #print("W2 dim:", W2.shape)
        #print("b2 dim:", b2.shape)
        #print("labels dim:", labels.shape)
                        
        ### YOUR CODE HERE - FORWARD PASS - compute cost with weight decay and store relevant values for backprop
        nll = np.zeros(X.shape[0])
        #Summy bois:
        W1_grad = np.zeros(W1.shape)
        W2_grad = np.zeros(W2.shape)
        b1_grad = np.zeros(b1.shape)
        b2_grad = np.zeros(b2.shape)
        lambd = c
        for i in range(X.shape[0]):
            #print("X: ", X)
            #print("X shape 0 ", X.shape[0])
            #Forward pass
            currentRow = X[i,:]
            #print("currentrow:", currentRow)
            #print("W1:", W1)
            a = currentRow @ W1
            #print("a:", a)
            b = a + b1
            c = relu(b)[0]
            d = c @ W2
            e = d + b2
            softmaxVec = softmax(e)[0]
            probability = softmaxVec[Y[i]]
            f = -np.log(probability)/X.shape[0]
            nll[i] = f
            
            #Backward pass
            
            df_de = (softmaxVec - labels[i,:])/X.shape[0]
            df_db2 = df_de
            df_dc = df_de @ W2.T
            print("c: ", c)
            print("c newaxis: ", c[:, np.newaxis])
            print("df_de:", df_de)
            df_dw2 = c[:, np.newaxis] @ df_de[:,np.newaxis].T 
            dc_db = np.diag(c > 0)
            df_db = df_dc @ dc_db
            df_da = df_db
            df_db1 = df_db
            df_dw1 = currentRow[:, np.newaxis] @ df_da[:,np.newaxis].T      

            W1_grad += df_dw1  
            W2_grad += df_dw2
            b1_grad += df_db1
            b2_grad += df_db2

        #print("lambd:", lambd)
        b2_grad = b2_grad 
        b1_grad = b1_grad 
        W1_grad = W1_grad
        W2_grad = W2_grad

        dweight_dw1 = lambd * 2 * W1
        dweight_dw2 = lambd * 2 * W2
        W1_grad += dweight_dw1
        W2_grad += dweight_dw2


        ### END CODE

        ### YOUR CODE HERE - BACKWARDS PASS - compute derivatives of all weights and bias, store them in d_w1, d_w2, d_b1, d_b2
        
            #nothin' to see here... *crickets*
        
        ### END CODE
        # the return signature
        loss = -np.mean(nll) + c * (np.sum(W1**2) + np.sum(W2**2))
        return loss, {'d_w1': W1_grad, 'd_w2': W2_grad, 'd_b1': b1_grad, 'd_b2': b2_grad}
        
    def fit(self, X_train, y_train, X_val, y_val, init_params, batch_size=32, lr=0.1, c=1e-4, epochs=30):
        """ Run Mini-Batch Gradient Descent on data X, Y to minimize the in sample error for Neural Net classification
        Printing the performance every epoch is a good idea to see if the algorithm is working
    
        Args:
           X_train: numpy array shape (n, d) - the training data each row is a data point
           y_train: numpy array shape (n,) int - training target labels numbers in {0, 1,..., k-1}
           X_val: numpy array shape (n, d) - the validation data each row is a data point
           y_val: numpy array shape (n,) int - validation target labels numbers in {0, 1,..., k-1}
           init_params: dict - has initial setting of parameters
           lr: scalar - initial learning rate
           batch_size: scalar - size of mini-batch
           c: scalar - weight decay parameter 
           epochs: scalar - number of iterations through the data to use

        Sets: 
           params: dict with keys {W1, W2, b1, b2} parameters for neural net
        returns
           hist: dict:{keys: train_loss, train_acc, val_loss, val_acc} each an np.array of size epochs of the the given cost after every epoch
           loss is the NLL loss and acc is accuracy
        """
        
        W1 = init_params['W1']
        b1 = init_params['b1']
        W2 = init_params['W2']
        b2 = init_params['b2']
        hist = {
            'train_loss': None,
            'train_acc': None,
            'val_loss': None,
            'val_acc': None, 
        }

        ### YOUR CODE HERE
        n = X_train.shape[0]
        train_loss = np.zeros[epochs]
        train_acc = np.zeros[epochs]
        val_loss = np.zeros[epochs]
        val_acc = np.zeros[epochs]
        params = {'W1': None, 'b1': None, 'W2': None, 'b2': None}
        for i in range(epochs):
            Xpermuted, Ypermuted = permute(X_train, y_train)
            batchesX = [Xpermuted[i:i+batch_size] for i in range(0, n, batch_size)]
            batchesY = [Ypermuted[i:i+batch_size] for i in range(0, n, batch_size)]
            for j in range(len(batchesX)):
                currentX = batchesX[j]
                currentY = batchesY[j]
                _, grad = self.cost_grad(currentX, currentY, c)
                W1 = W1 - lr * grad['d_w1']
                b1 = b1 - lr * grad['d_b1']
                W2 = W2 - lr * grad['d_w2']
                b2 = b2 - lr * grad['d_b2']

            params = {'W1': W1, 'b1': b1, 'W2': W2, 'b2': b2}
            
            costTrain, _ =  self.cost_grad(X_train, y_train, c)
            train_loss[i] = costTrain 
            scoreTrain = self.score(X_train, y_train, params)   
            train_acc[i] = scoreTrain
            
            costVal, _ =  self.cost_grad(X_val, y_val)
            val_loss[i] = costVal
            scoreVal = self.score(X_val, y_val, params)
            val_acc[i] = scoreVal
            
            print("Epoch ", i)
            print("Train_loss", costTrain)
            print("Train_acc", scoreTrain)
            print("Val_loss", costVal)
            print("Val_acc", scoreVal)
        
        hist = { 
            'train_loss': train_loss,
            'train_acc': train_acc,
            'val_loss': val_loss,
            'val_acc': val_acc, 
        }
        self.params = params
        
        ### END CODE
        # hist dict should look like this with something different than none
        #hist = {'train_loss': None, 'train_acc': None, 'val_loss': None, 'val_acc': None}
        ## self.params should look like this with something better than none, i.e. the best parameters found.
        # self.params = {'W1': None, 'b1': None, 'W2': None, 'b2': None}
        return hist
        
def permute(X, y):
    assert y.shape[0] == X.shape[0]
    xy = np.hstack((X,y[:, np.newaxis]))
    perm = np.random.permutation(xy)
    perm_y = perm[:,-1]
    perm_x = perm[:,:-1]
    assert y.shape == perm_y.shape
    assert X.shape == perm_x.shape
    return perm_x, perm_y


def numerical_grad_check(f, x, key):
    """ Numerical Gradient Checker """
    eps = 1e-6
    h = 1e-5
    # d = x.shape[0]
    cost, grad = f(x)
    grad = grad[key]
    num_grads = []
    print("Grad: ", grad)
    it = np.nditer(x, flags=['multi_index'])
    while not it.finished:    
        dim = it.multi_index    
        #print(dim)
        tmp = x[dim]
        x[dim] = tmp + h
        cplus, _ = f(x)
        x[dim] = tmp - h 
        cminus, _ = f(x)
        x[dim] = tmp
        num_grad = (cplus-cminus)/(2*h)
        print('cplus cminus', cplus, cminus, cplus-cminus)
        print('dim, grad, num_grad, grad-num_grad', dim, grad[dim], num_grad, grad[dim]-num_grad)
        assert np.abs(num_grad - grad[dim]) < eps, 'numerical gradient error index {0}, numerical gradient {1}, computed gradient {2}'.format(dim, num_grad, grad[dim])
        num_grads.append(num_grad)
        it.iternext()
    print("num_grads:", num_grads)

def test_grad():
    stars = '*'*5
    print(stars, 'Testing  Cost and Gradient Together')
    input_dim = 7
    hidden_size = 1
    output_size = 3
    nc = NetClassifier()
    params = get_init_params(input_dim, hidden_size, output_size)

    nc = NetClassifier()
    X = np.random.randn(7, input_dim)
    y = np.array([0, 1, 2, 0, 1, 2, 0])

    f = lambda z: nc.cost_grad(X, y, params, c=1.0)
    print('\n', stars, 'Test Cost and Gradient of b2', stars)
    numerical_grad_check(f, params['b2'], 'd_b2')
    print(stars, 'Test Success', stars)
    
    print('\n', stars, 'Test Cost and Gradient of w2', stars)
    numerical_grad_check(f, params['W2'], 'd_w2')
    print('Test Success')
    
    print('\n', stars, 'Test Cost and Gradient of b1', stars)
    numerical_grad_check(f, params['b1'], 'd_b1')
    print('Test Success')
    
    print('\n', stars, 'Test Cost and Gradient of w1', stars)
    numerical_grad_check(f, params['W1'], 'd_w1')
    print('Test Success')

if __name__ == '__main__':
    
    '''input_dim = 3
    hidden_size = 5
    output_size = 4
    batch_size = 1
    nc = NetClassifier()
    params = {'W1': np.ones([3,5]), 'b1': np.ones([1,5]), 'W2': np.ones([5,4]), 'b2':np.ones([1,4]) }
    X = np.array([[1,2,3]])
    Y = np.array([2])
    #loss, newParams = nc.cost_grad(X,Y,params,c=0)
    #nc.cost_grad(X, Y, params, c=0)   
    #test_grad()

    stars = '*'*5
    f = lambda z: nc.cost_grad(X, Y, params, c=1.0)
    print('\n', stars, 'Test Cost and Gradient of b2', stars)
    numerical_grad_check(f, params['b2'], 'd_b2')
    print(stars, 'Test Success', stars)
    
    print('\n', stars, 'Test Cost and Gradient of w2', stars)
    numerical_grad_check(f, params['W2'], 'd_w2')
    print('Test Success')
    
    print('\n', stars, 'Test Cost and Gradient of b1', stars)
    numerical_grad_check(f, params['b1'], 'd_b1')
    print('Test Success')
    
    print('\n', stars, 'Test Cost and Gradient of w1', stars)
    numerical_grad_check(f, params['W1'], 'd_w1')
    print('Test Success')

    print("params", params)
    print("-----------------------------------")
    print("loss", loss)
    print("gradient W1", newParams['d_w1'])
    print("gradient W2", newParams['d_w2'])
    print("gradient b1", newParams['d_b1'])
    print("gradient b2", newParams['d_b2'])'''
    
    
    
    
    
    input_dim = 3
    hidden_size = 5
    output_size = 4
    batch_size = 7
    nc = NetClassifier()
    params = get_init_params(input_dim, hidden_size, output_size)
    X = np.random.randn(batch_size, input_dim)
    Y = np.array([0, 1, 2, 0, 1, 2, 0])
    
    
    nc.cost_grad(X, Y, params, c=0)   
    test_grad()
