<?php

namespace App\Jobs\Render;

use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Queue\Middleware\ThrottlesExceptionsWithRedis;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;
use Carbon\Carbon;

use App\Constants\Thumbnails\ThumbnailType;
use App\Contracts\Models\IThumbnailable;
use App\Models\Item\Item;
use App\Models\User\User;
use App\Models\Polymorphic\Thumbnail;

class RenderThumbnail implements ShouldQueue, ShouldBeUnique
{
    use Dispatchable, InteractsWithQueue, Queueable;

    protected IThumbnailable $model;
    protected ThumbnailType $type;
    public $uniqueFor = 900;

    public function uniqueId(): string
    {
        return $this->model->getMorphClass() . ':' . $this->model->id . ":" . $this->type->value;
    }

    public function middleware(): array
    {
        return [(new ThrottlesExceptionsWithRedis(2, 3))->backoff(1)];
    }

    public function retryUntil(): \DateTime
    {
        return now()->addMinutes(5);
    }

    public function __construct(IThumbnailable $model, ThumbnailType $type)
    {
        $this->model = $model;
        $this->type = $type;
    }

    public function handle()
    {
        // Generate a new unique UUID for the thumbnail filename.
        $newThumbnailUuid = Str::uuid()->toString();
        
        $renderParams = [];

        // Build the parameters based on the model type.
        if ($this->model instanceof User) {
            /** @var User $user */
            $user = $this->model;
            $avatar = $user->avatar;
            if (!$avatar) {
                Log::warning('RenderThumbnail job: No avatar for user.', ['user_id' => $user->id]);
                return;
            }

            $items = $avatar->items;
            $colors = $avatar->colors;
            $hats = $items->get('hats', [0,0,0,0,0,0]);

            $renderParams = [
                'RenderType' => 'user',
                'hash' => $newThumbnailUuid,
                'head_color' => $colors->get('head', 'f3b700'),
                'torso_color' => $colors->get('torso', 'f3b700'),
                'leftLeg_color' => $colors->get('left_leg', 'f3b700'),
                'rightLeg_color' => $colors->get('right_leg', 'f3b700'),
                'leftArm_color' => $colors->get('left_arm', 'f3b700'),
                'rightArm_color' => $colors->get('right_arm', 'f3b700'),
                'face' => $items->get('face', 0) ?: 'none',
                'tool' => $items->get('tool', 0) ?: 'none',
                'shirt' => $items->get('shirt', 0) ?: 'none',
                'pants' => $items->get('pants', 0) ?: 'none',
                'tshirt' => $items->get('tshirt', 0) ?: 'none',
                'hat_1' => ($hats[0] ?? 0) ?: 'none',
                'hat_2' => ($hats[1] ?? 0) ?: 'none',
                'hat_3' => ($hats[2] ?? 0) ?: 'none',
                'hat_4' => ($hats[3] ?? 0) ?: 'none',
                'hat_5' => ($hats[4] ?? 0) ?: 'none',
                'hat_6' => ($hats[5] ?? 0) ?: 'none',
            ];
        } else if ($this->model instanceof Item) {
            /** @var Item $item */
            $item = $this->model;
            $renderParams = [
                'RenderType' => 'item',
                'item' => $item->id,
                'itemhash' => $newThumbnailUuid, 
                'itemtype' => $item->type,
            ];
        } else {
            Log::warning('RenderThumbnail job dispatched for unsupported model.', ['model_type' => get_class($this->model)]);
            return;
        }

        $renderUrl = config('app.render_server_url');
        if (empty($renderUrl)) {
            Log::error('Render server URL is not configured.');
            return;
        }

        try {
            $response = Http::withHeaders(['Aeo-Access-Key' => config('app.render_server_key')])
                ->timeout(60)
                ->get($renderUrl, $renderParams);

                // If rendering was successful, update the model's hash and create the thumbnail record.

                // Create a new thumbnail record to signify completion.
                $thumb = Thumbnail::create([
                    'uuid' => $newThumbnailUuid, // Store the UUID here for reference.
                    'contents_uuid' => $newThumbnailUuid, // Using the same UUID for content tracking.
                    'expires_at' => Carbon::now()->addYear(),
                ]);
                
                // Detach any old thumbnails of the same type and attach the new one.
                $this->model->thumbnails()->wherePivot('thumbnail_type', $this->type->value)->detach();
                $this->model->thumbnails()->attach($thumb, ['thumbnail_type' => $this->type]);

                Log::info('Successfully rendered and saved thumbnail.', ['model' => get_class($this->model), 'id' => $this->model->id, 'hash' => $newThumbnailUuid]);

        } catch (\Exception $e) {
            Log::error('Failed to connect to render service.', ['error' => $e->getMessage(), 'model' => get_class($this->model), 'id' => $this->model->id]);
            $this->release(60);
        }
    }
}
