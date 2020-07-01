# Go-Bumblebee-Jazz

Go-Bumblebee-Jazz will setup CodeBuild, ECR and GitHub as upstream for release pipeline of Go-Bumblebee. 

## High-level Flow

|-------------------------------------------------------------------------------------------|

Go-Bumblebee-Jazz at GitHub -> CodeBuild -> ECR                             -> ArgoCD -> ...
                                         -> Go-Bumblebee-Jazz at GitHub

|---------------------- CI -----------------------------------------------|------- CD ------|