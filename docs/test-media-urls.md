# 🔗 URLs Válidas para Testes de Mídia

## 📸 **Imagens (Funcionais)**

### JPG/JPEG
```
https://picsum.photos/800/600.jpg
https://picsum.photos/400/300.jpg
https://via.placeholder.com/500x300.jpg
```

### PNG
```
https://via.placeholder.com/400x300.png
https://picsum.photos/600/400.png
```

### WebP
```
https://developers.google.com/speed/webp/gallery/1.webp
https://developers.google.com/speed/webp/gallery/2.webp
```

## 🎵 **Áudio (Funcionais)**

### WAV
```
https://www.soundjay.com/misc/sounds/bell-ringing-05.wav
https://samplelib.com/lib/preview/wav/sample-3s.wav
```

### MP3
```
https://www.learningcontainer.com/wp-content/uploads/2020/02/Kalimba.mp3
https://samplelib.com/lib/preview/mp3/sample-3s.mp3
```

### OGG
```
https://samplelib.com/lib/preview/ogg/sample-3s.ogg
```

## 🎬 **Vídeo (Funcionais)**

### MP4
```
https://sample-videos.com/zip/10/mp4/SampleVideo_360x240_1mb.mp4
https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4
https://www.learningcontainer.com/wp-content/uploads/2020/05/sample-mp4-file.mp4
```

### WebM
```
https://sample-videos.com/zip/10/webm/SampleVideo_360x240_1mb.webm
```

## 📄 **Documentos (Funcionais)**

### PDF
```
https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf
https://www.adobe.com/support/products/enterprise/knowledgecenter/media/c4611_sample_explain.pdf
```

### TXT
```
https://www.learningcontainer.com/wp-content/uploads/2020/04/sample-text-file.txt
```

### ZIP
```
https://www.learningcontainer.com/wp-content/uploads/2020/05/sample-zip-file.zip
```

## 🎭 **Stickers (WebP)**

### Sticker WebP
```
https://developers.google.com/speed/webp/gallery/1.webp
https://developers.google.com/speed/webp/gallery/2.webp
```

## 🔧 **URLs de Backup (GitHub Raw)**

### Imagens
```
https://raw.githubusercontent.com/microsoft/vscode/main/resources/linux/code.png
https://raw.githubusercontent.com/github/explore/main/topics/javascript/javascript.png
```

### Documentos
```
https://raw.githubusercontent.com/microsoft/vscode/main/README.md
https://raw.githubusercontent.com/microsoft/vscode/main/LICENSE.txt
```

## ⚠️ **URLs Problemáticas (Evitar)**

### Não Funcionam
```
❌ https://file-examples.com/* (403 Forbidden)
❌ https://sample-videos.com/* (Timeout/404)
❌ https://www.gstatic.com/* (404)
```

## 🎯 **URLs Recomendadas para Testes**

### Para Imagens
- **Pequenas**: `https://via.placeholder.com/300x200.jpg`
- **Médias**: `https://picsum.photos/800/600.jpg`
- **PNG**: `https://via.placeholder.com/400x300.png`

### Para Áudio
- **WAV**: `https://www.soundjay.com/misc/sounds/bell-ringing-05.wav`
- **MP3**: `https://www.learningcontainer.com/wp-content/uploads/2020/02/Kalimba.mp3`

### Para Vídeo
- **MP4 Pequeno**: `https://sample-videos.com/zip/10/mp4/SampleVideo_360x240_1mb.mp4`
- **MP4 Grande**: `https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4`

### Para Documentos
- **PDF**: `https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf`
- **TXT**: `https://www.learningcontainer.com/wp-content/uploads/2020/04/sample-text-file.txt`

## 📊 **Status de Confiabilidade**

| Serviço | Confiabilidade | Formatos | Observações |
|---------|----------------|----------|-------------|
| picsum.photos | ✅ Alta | JPG, PNG | Sempre funciona |
| via.placeholder.com | ✅ Alta | JPG, PNG | Sempre funciona |
| soundjay.com | ✅ Alta | WAV | Sempre funciona |
| learningcontainer.com | ✅ Alta | MP3, TXT | Sempre funciona |
| w3.org | ✅ Alta | PDF | Sempre funciona |
| sample-videos.com | ⚠️ Média | MP4, WebM | Às vezes timeout |
| file-examples.com | ❌ Baixa | Vários | Frequente 403 |
