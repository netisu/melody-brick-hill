<?php

namespace App\Http\Controllers\API;

use App\Http\Controllers\Controller;
use App\Http\Requests\Shop\Preview;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;
use Intervention\Image\ImageManagerStatic as Image;

class RenderAPIController extends Controller
{
    /**
     * Generates a preview for a new item by saving temporary assets
     * and calling the Go rendering service.
     *
     * @param \App\Http\Requests\Shop\Preview $request
     * @return \Illuminate\Http\Response
     */
    public function newPreview(Preview $request)
    {
        // Generate a unique temporary hash for this preview session.
        $tempHash = 'preview_' . Str::uuid();

        // Use the 'public' disk. This must be configured to point to the web-accessible
        // directory that the Go service uses as its `cdnDirectory`.
        $uploadDisk = Storage::disk('public');
        
        $texturePath = "uploads/{$tempHash}.png";
        $meshPath = "uploads/{$tempHash}.obj";
        $previewImagePath = "thumbnails/{$tempHash}.png";

        try {
            // Save uploaded texture file, if provided.
            if ($request->hasFile('texture')) {
                $textureContent = Image::make($request->file('texture'))->encode('png');
                $uploadDisk->put($texturePath, $textureContent);
            }

            // Save uploaded mesh file, if provided.
            if ($request->hasFile('mesh')) {
                $meshContent = file_get_contents($request->file('mesh')->getRealPath());
                $uploadDisk->put($meshPath, $meshContent);
            }

            // Call the Go rendering service.
            $renderUrl = config('app.render_server_url');
            if (empty($renderUrl)) {
                throw new \Exception('Render service URL is not configured.');
            }
            
$response = Http::withHeaders([
    'Aeo-Access-Key' => config('app.render_server_key')
])->timeout(60)->post($renderUrl, [
    'RenderType' => 'item',
    'item' => $tempHash,
    'itemtype' => $request->type,
]);

            if ($response->failed()) {
                Log::error('Render service failed for preview.', ['hash' => $tempHash, 'status' => $response->status(), 'body' => $response->body()]);
                throw new \Exception('Failed to generate item preview from render service.');
            }

            // The Go service saves the file, so we read it from the shared disk.
            if (!$uploadDisk->exists($previewImagePath)) {
                Log::error('Render service succeeded but preview image not found on disk.', ['path' => $previewImagePath]);
                throw new \Exception('Failed to retrieve generated preview image.');
            }

            $previewImageContent = $uploadDisk->get($previewImagePath);

            // Return the final image as a data URL.
            return Image::make($previewImageContent)
                ->resize(256, 256)
                ->encode('webp')
                ->response('data-url');

        } finally {
            // Clean up all temporary files regardless of success or failure.
            if ($uploadDisk->exists($texturePath)) $uploadDisk->delete($texturePath);
            if ($uploadDisk->exists($meshPath)) $uploadDisk->delete($meshPath);
            if ($uploadDisk->exists($previewImagePath)) $uploadDisk->delete($previewImagePath);
        }
    }
}
