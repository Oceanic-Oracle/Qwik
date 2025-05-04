namespace Qwik;

public partial class ProfilePage : ContentPage
{
    string Login = string.Empty;
    string Email = string.Empty;
    public ProfilePage()
	{
		InitializeComponent();
        Login = "sdvsb";
        Email = "sdbsdb";
    }

    protected override async void OnAppearing()
	{
        if (string.IsNullOrEmpty(Config.JWT))
        {
            await Shell.Current.Navigation.PushAsync(new AuthPage());
            return;
        }
    }

    private void OnSearchClicked(object sender, EventArgs e)
    {
        
    }

    private async void OpenProfileClicked(object sender, EventArgs e)
    {
        string action = await DisplayActionSheet(
        $"�����: {Login}{Environment.NewLine}�����: {Email}",
        "������", 
        "�����");

        if (action == "�����")
        {
            Config.JWT = "";
            await Shell.Current.Navigation.PushAsync(new AuthPage());
        }
    }
}