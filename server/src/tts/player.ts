import { Howl } from "howler";

class AudioManager {
  private current: Howl | null = null;

  playBase64Mp3(base64: string) {
    this.stop();

    this.current = new Howl({
      src: [`data:audio/mp3;base64,${base64}`],
      format: ["mp3"],
      html5: true, // important for larger audio files
    });

    this.current.play();
  }

  stop() {
    if (this.current) {
      this.current.stop();
      this.current.unload();
      this.current = null;
    }
  }
}

export const audioManager = new AudioManager();
