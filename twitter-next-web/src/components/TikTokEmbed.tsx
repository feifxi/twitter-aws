export function TikTokEmbed() {
  return (
    <div className="flex items-center justify-center w-full h-full">
        <iframe 
            src="https://www.tiktok.com/embed/v2/7606653926787566864"
            className="w-full h-full border-0"
            allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
            sandbox="allow-scripts allow-same-origin allow-popups allow-presentation"
            style={{ maxWidth: '105px', minWidth: '325px', height: '100%', maxHeight: '725px', borderRadius: '12px' }}
        ></iframe>
    </div>
  );
}
