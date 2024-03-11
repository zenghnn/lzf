# lzf
A rewrite of the C# version of Chase Pettit's lzf with Go

## lzf 算法的go实现

* 在unity离线日志并同步功能的过程中,为了在定时给服务器发送日志文件,需要把内容压缩,否则会浪费很大的资源.所以想使用gzip.但是在使用时.Unity.IO.Compression并不理想.包体大,性能开销较大.代码复杂.最重要的是,和go自带的gzip解压缩并不对应.解压缩出来的内容并不正确.所以需要重选一个压缩算法.
* 于是选到了lzf.代码简单,而且没有其他依赖.单个文件就可以实现功能.但新的问题又出现了.网上提供的[c#](https://github.com/Chaser324/LZF/blob/master/CLZF2.cs)和[go版本](https://github.com/tav/golly/lzf)并不对应.go解压缩出来的内容不正确.
* 最后我只好花了半天时间把c#版本用go 重写.并进行了测试.
* 希望对大家会有用
