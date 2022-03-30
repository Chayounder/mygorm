Gorm programming practices

echo "# mygorm" >> README.md
git init
git add README.md
git commit -m "first commit"
git branch -M main
git remote add origin https://github.com/Chayounder/mygorm.git
git push -u origin main



$ ssh-keygen -t rsa -C "2442512720@qq.com"
Generating public/private rsa key pair.
Enter file in which to save the key (/c/Users/l0042884/.ssh/id_rsa): /c/Users/l0042884/.ssh/github_rsa
Enter passphrase (empty for no passphrase):
Enter same passphrase again:

#如果系统已经有ssh-key 代理 ,执行下面的命令可以删除
$ ssh-add -D
2048 SHA256:LC6iL7q5LMX1fvTBauKyqMJQeouIpowXoGvVs3pHLYg 2442512720@qq.com (RSA)

#将私钥添加到 ssh-agent
$ ssh-add /c/Users/l0042884/.ssh/github_rsa
Could not open a connection to your authentication agent.

#如果出现以上错误，执行如下指令
$ ssh-agent bash

$ ssh-add /c/Users/l0042884/.ssh/github_rsa
Identity added: /c/Users/l0042884/.ssh/github_rsa (2442512720@qq.com)

$ ssh-add -l
2048 SHA256:LC6iL7q5LMX1fvTBauKyqMJQeouIpowXoGvVs3pHLYg 2442512720@qq.com (RSA)
