t = int(input())
for _ in range(t):
    n = int(input())
    p = list(map(int,input().split()))

    count = [0]

    def check(left, right):
        
        if right - left == 1:
            return p[left],p[left]
             
        mid = (left+ right)//2
        min_left,max_left = check(left,mid)
        min_right,max_right = check(mid, right)
        if min_left == -1 or min_right == -1:
            return -1,-1
        if max_left < min_right:
            return min_left,max_right
        elif max_right < min_left:
            count[0]+=1
            return min_right,max_left
        else:
            return -1,-1


    res = check(0,n)

    if res[0] == -1:
        print("-1")
    else:
        print(count[0])


    
