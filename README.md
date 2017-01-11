
# lightweight MP4-parser

:construction:

## Usage
    ```go
    f, _ := os.Open(`X:\file.mp4`)
	
	p := mp4parser.NewParser(f)
	info, _ := p.Parse()
    
    f.Close()
	fmt.Println(info)
    ```
