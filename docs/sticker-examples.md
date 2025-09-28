# 🎭 Exemplos de Stickers para Teste

## 📋 **Especificações WhatsApp Stickers**

| Requisito | Valor | Implementado |
|-----------|-------|--------------|
| Formato | WebP | ✅ |
| Dimensões recomendadas | 512x512px | ⚠️ (WhatsApp faz resize) |
| Tamanho máximo (estático) | 100KB | ✅ |
| Tamanho máximo (animado) | 500KB | ✅ |
| MIME type | image/webp | ✅ |

## 🎯 **Sticker Base64 (Pequeno - ~2KB)**

### Sticker WebP Simples (16x16px)
```
data:image/webp;base64,UklGRlIAAABXRUJQVlA4IEYAAAAwAQCdASoQABAAAgA0JaQAA3AA/vuqAAA=
```

## 🔗 **URLs Válidas para Teste**

### URLs Potencialmente Funcionais
```
# WebP pequenos
https://www.gstatic.com/hostedimg/webp_logo_small.webp
https://developers.google.com/speed/webp/gallery/1.webp

# GitHub Raw (se existir)
https://raw.githubusercontent.com/webmproject/libwebp/main/examples/1.webp
```