const escapeCsvValue = (value) => {
  const text = String(value ?? '');
  return `"${text.replaceAll('"', '""')}"`;
};

export const downloadCsv = (filename, headers, rows) => {
  const csv = [headers, ...rows]
    .map((row) => row.map(escapeCsvValue).join(','))
    .join('\r\n');
  const blob = new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
};

export const formatPublicPollUrl = (pollId) => `${window.location.origin}/p/${pollId}`;

export const downloadQrCode = (url, filename = 'votepulse-qr.png') => {
  const image = new Image();
  image.crossOrigin = 'anonymous';
  image.onload = () => {
    const canvas = document.createElement('canvas');
    canvas.width = image.naturalWidth;
    canvas.height = image.naturalHeight;
    canvas.getContext('2d').drawImage(image, 0, 0);
    const anchor = document.createElement('a');
    anchor.href = canvas.toDataURL('image/png');
    anchor.download = filename;
    anchor.click();
  };
  image.src = `https://quickchart.io/qr?text=${encodeURIComponent(url)}&size=640`;
};

export const shareOrCopy = async (title, url) => {
  if (navigator.share) {
    try {
      await navigator.share({ title, url });
    } catch (error) {
      if (error?.name !== 'AbortError') throw error;
      return 'dismissed';
    }
    return 'shared';
  }
  await navigator.clipboard.writeText(url);
  return 'copied';
};
