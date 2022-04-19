# idea/IntelliJ launcher

###### create a temp file under ~/Documents/new and open in IntelliJ
./i

###### open new/existing file in IntelliJ. if this dir or one of the parent directories has an idea project open that project as well
./i my-file

###### open idea project from this dir or one of it's parent directories if found; otherwise create a new idea project
./i my-dir

###### pipe stdin into a temp file
kubectl logs -f pod | i

###### pipe stdin into a new/existing file
kubectl logs -f pod | i my-file