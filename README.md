# excel-to-json
将超大excel转化为json的工具

> 使用方法
> 
执行命令：`go run main.go --help`  获取帮助

>Flags:
> 
>     --[no-]help      Show context-sensitive help (also try --help-long and --help-man)
>     -i, --id=ID          文件相同属性
>     -s, --source=SOURCE  被分割文件地址(绝对地址)
>     -d, --dest="/home/ipcc/data/crm-import-tmp" 分割文件存储地址,不带`/`是程序运行的相对地址
>     -c, --num=20         并发数
>     -p, --handle=10000   每个文件的数据量
>     -f, --[no-]debug     开启日志

执行启动：`go run main.go -i 'res' -d 'excelTmp' -s 'book.xlsx' -c 2 -p 5 -f`  

> 结果演示

![image](https://github.com/Alke-meng/excel-to-json/images/1.png)

![image](https://github.com/Alke-meng/excel-to-json/images/2.png)

